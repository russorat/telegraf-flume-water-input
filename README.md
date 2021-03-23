# Flume Water Input Plugin

This plugin will query the Flume Water API to fetch usage data about your account.

Note: Currently personal clients are allowed 120 requests per hour to the API. Once you have gone over the limit you will get a 429 rate limit error.

Telegraf minimum version: Telegraf 18.0
Plugin minimum tested version: 18.0

### Build and Run

To build this plugin, just run:

```sh
make
```

Which will build the binary `./bin/flume-water`

You can run it with `./bin/flume-water --config plugin.conf`

### Configuration

This is the plugin configuration that is expected via the `--config` param

```toml
[[inputs.flume_water]]
  client_id = "clientid"
    client_secret = "secret"
	username = "username"
	password = "password"
	## If this isn't set, we will fetch your device list and pick the first one
	#device_id = ""
	## lookback_mins is the amount of minutes to look back when querying data. This helps catch any late arriving data
	#lookback_mins = 5
	## units can be one of GALLONS, LITERS, CUBIC_FEET, or CUBIC_METERS
	#units = "GALLONS"
```

And this is an example of how to configure this with the Telegraf execd plugin that you would run with `telegraf --config telegraf.conf`

```toml
[[inputs.execd]]
  ## One program to run as daemon.
  ## NOTE: process and each argument should each be their own string
  command = ["/path/to/flume-water", "--config", "/path/to/plugin.conf"]

  ## Define how the process is signaled on each collection interval.
  ## Valid values are:
  ##   "none"    : Do not signal anything. (Recommended for service inputs)
  ##               The process must output metrics by itself.
  ##   "STDIN"   : Send a newline on STDIN. (Recommended for gather inputs)
  ##   "SIGHUP"  : Send a HUP signal. Not available on Windows. (not recommended)
  ##   "SIGUSR1" : Send a USR1 signal. Not available on Windows.
  ##   "SIGUSR2" : Send a USR2 signal. Not available on Windows.
  signal = "none"

  ## Delay before the process is restarted after an unexpected termination
  restart_delay = "10s"

  ## Data format to consume.
  ## Each data format has its own unique set of configuration options, read
  ## more about them here:
  ## https://github.com/influxdata/telegraf/blob/master/docs/DATA_FORMATS_INPUT.md
  data_format = "influx"
  ```

### Polling

By default, the plugin is designed to fetch data every minute. Using the example config above doesn't care how often Telegraf's inverval is set.

To customize, you can send in `--poll_interval 1h` which will fetch query results every hour.

To allow Telegraf to control the interval of the gather, set `--poll_interval_disabled` on the `flume-water` command and `signal = "STDIN"` in your telegraf `execd` config.

### Metrics 

- flume_water
  - tags:
    - bridge_id
    - device_id
    - device_type
    - location_building_type
    - location_city
    - location_name
    - location_postal_code
    - location_state
    - request_id
    - user_email
  - fields:
    - gallons (float)

### Example Output

```
flume_water,bridge_id=45645645645634,device_id=34534534656456,device_type=2,location_building_type=SINGLE_FAMILY_HOME,location_city=San\ Francisco,location_name=Home,location_postal_code=94110,location_state=CA,request_id=flume-water-telegraf-input,user_email=russ@example.com gallons=2.324 1616069940000000000
```
