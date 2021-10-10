module github.com/e-dard/influxdb-iox-grafana

go 1.16

require (
	github.com/apache/arrow/go/arrow v0.0.0-20210223225224-5bea62493d91
	github.com/golang/protobuf v1.5.2 // indirect
	github.com/grafana/grafana-plugin-sdk-go v0.114.0
	github.com/grafana/grafana-starter-datasource-backend v0.0.0-20210820074628-c40d1723f2f5
	google.golang.org/genproto v0.0.0-20200911024640-645f7a48b24f
	google.golang.org/grpc v1.37.1
	google.golang.org/protobuf v1.27.1
)

replace github.com/grafana/grafana-plugin-sdk-go => github.com/e-dard/grafana-plugin-sdk-go v0.114.1-0.20211010131548-afb84c377f7c
