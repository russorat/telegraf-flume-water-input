package flumewater

import (
	"fmt"
	"time"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/metric"
	"github.com/influxdata/telegraf/plugins/inputs"
	flume "github.com/russorat/flume-water-go-client"
)

const (
	metricName           = "flume_water"
	defaultClientTimeout = 5 * time.Second
)
const MetricName = "flume_water"

type FlumeWater struct {
	Timeout      time.Duration `toml:"timeout"`
	ClientID     string        `toml:"client_id"`
	ClientSecret string        `toml:"client_secret"`
	Username     string        `toml:"username"`
	Password     string        `toml:"password"`
	DeviceID     string        `toml:"device_id"`

	Queries []flume.FlumeWaterQuery `toml:"query"`

	client flume.Client
	device flume.FlumeWaterDevice
	Log    telegraf.Logger `toml:"-"`
}

type FlumeQueryRequest struct {
	Bucket          string `json:"bucket" toml:"bucket"`
	GroupMultiplier int    `json:"group_multiplier,omitempty" toml:"group_multiplier"`
	SinceDatetime   string `json:"since_datetime" toml:"since_datetime"`
	UntilDatetime   string `json:"until_datetime,omitempty" toml:"until_datetime"`
	Operation       string `json:"operation,omitempty" toml:"operation"`
	Units           string `json:"units,omitempty" toml:"units"`
	SortDirection   string `json:"sort_direction,omitempty" toml:"sort_direction"`
	RequestID       string `json:"request_id" toml:"request_id"`
}

func init() {
	inputs.Add("flume_water", func() telegraf.Input {
		return &FlumeWater{
			Timeout: defaultClientTimeout,
		}
	})
}

func (fw *FlumeWater) SampleConfig() string {
	return `
    ## Timeout for HTTP message
    # timeout = "5s"
    
    client_id = "clientid"
    client_secret = "secret"
	username = "username"
	password = "password"
	## If this isn't set, we will fetch your device list and pick one
	#device_id = ""
`
}

func (fw *FlumeWater) Description() string {
	return "Gathers metrics from Flume Water Meter API"
}

func (fw *FlumeWater) Gather(a telegraf.Accumulator) error {
	fw.client = flume.NewClient(fw.ClientID, fw.ClientSecret, fw.Username, fw.Password)
	if fw.device.ID == "" {
		var err error
		fw.device, err = fw.client.FetchUserDevice(fw.DeviceID, flume.FlumeWaterFetchDeviceRequest{IncludeUser: true, IncludeLocation: true})
		if err != nil {
			a.AddError(err)
		}
	}
	until := time.Now()
	since := time.Now()
	since = since.Add(-5 * time.Minute)

	values := flume.FlumeWaterQueryRequest{
		Queries: []flume.FlumeWaterQuery{
			{
				Bucket:        flume.FlumeWaterBucketMinute,
				SinceDatetime: since.Format("2006-01-02 15:04") + ":00",
				UntilDatetime: until.Format("2006-01-02 15:04") + ":00",
				RequestID:     "flume-water-telegraf-input",
				Units:         flume.FlumeWaterUnitGallon,
			}},
	}
	results, err := fw.client.QueryUserDevice(fw.DeviceID, values)
	if err != nil {
		a.AddError(err)
	}
	fw.sendMetric(a, &results)
	return nil
}

func (fw *FlumeWater) Stop() {
	fw.client.Close()
}

func (fw *FlumeWater) sendMetric(a telegraf.Accumulator, results *[]flume.FlumeWaterQueryResult) {
	for _, s := range *results {
		for key, element := range s {
			for _, bucket := range element {
				f := map[string]interface{}{
					"gallons": bucket.Value,
				}
				t := map[string]string{
					"request_id":             key,
					"device_id":              fw.device.ID,
					"bridge_id":              fw.device.BridgeID,
					"device_name":            fw.device.Name,
					"device_type":            fmt.Sprint(fw.device.Type),
					"user_email":             fw.device.User.EmailAddress,
					"location_name":          fw.device.Location.Name,
					"location_city":          fw.device.Location.City,
					"location_state":         fw.device.Location.State,
					"location_postal_code":   fw.device.Location.PostalCode,
					"location_building_type": fw.device.Location.BuildingType,
				}
				tz, err := time.LoadLocation(fw.device.Location.TZ)
				if err != nil {
					a.AddError(err)
				}

				dt, err := time.ParseInLocation("2006-01-02 15:04:05", bucket.Datetime, tz)
				if err != nil {
					a.AddError(err)
				}
				m, _ := metric.New(metricName, t, f, dt)
				a.AddMetric(m)
			}

		}
	}
}
