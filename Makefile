.DEFAULT_GOAL := build

build:
	go build -o bin/flume-water cmd/main.go

deps:
	go mod vendor

clean:
	rm -r bin

test:
	go test -timeout 30s -count=1 ./plugins/inputs/flume-water

run:
	make build
	./bin/flume-water --config plugin.conf --poll_interval 5s
