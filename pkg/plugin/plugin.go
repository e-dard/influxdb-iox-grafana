package plugin

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/url"
	"time"

	pb "github.com/e-dard/influxdb-iox-grafana/pkg/iox/github.com/influxdata/iox/management/v1"
	"github.com/grafana/grafana-plugin-sdk-go/backend"
	"github.com/grafana/grafana-plugin-sdk-go/backend/instancemgmt"
	"github.com/grafana/grafana-plugin-sdk-go/backend/log"
	"google.golang.org/grpc"
	"google.golang.org/grpc/status"
)

var (
	_ backend.QueryDataHandler      = (*IOxDatasource)(nil)
	_ backend.CheckHealthHandler    = (*IOxDatasource)(nil)
	_ backend.StreamHandler         = (*IOxDatasource)(nil)
	_ instancemgmt.InstanceDisposer = (*IOxDatasource)(nil)
)

type datasourceModel struct {
	Host     string `json:"host"`
	Database string `json:"database"`
}

// IOxDatasource can respond to queries and describe its health.
type IOxDatasource struct {
	Host     string
	Database string

	mgt_client   pb.ManagementServiceClient
	query_client *FlightClient
	err          error // error, if any, returned when initialising client(s)
}

// NewIOxDatasource creates a new datasource instance.
func NewIOxDatasource(s backend.DataSourceInstanceSettings) (instancemgmt.Instance, error) {
	var dm datasourceModel
	if err := json.Unmarshal(s.JSONData, &dm); err != nil {
		return nil, err
	}

	ds := &IOxDatasource{Host: dm.Host, Database: dm.Database}
	ds.err = ds.connect() // initialise underlying connection to IOx
	return ds, nil
}

// Dispose here tells plugin SDK that plugin wants to clean up resources when a new instance
// created. As soon as datasource settings change detected by SDK old datasource instance will
// be disposed and a new one will be created using NewIOxDatasource factory function.
func (d *IOxDatasource) Dispose() {
	// Clean up datasource instance resources.
}

// QueryData handles multiple queries and returns multiple responses.
// req contains the queries []DataQuery (where each query contains RefID as a unique identifier).
// The QueryDataResponse contains a map of RefID to the response for each query, and each response
// contains Frames ([]*Frame).
func (d *IOxDatasource) QueryData(ctx context.Context, req *backend.QueryDataRequest) (*backend.QueryDataResponse, error) {
	log.DefaultLogger.Debug("QueryData called", "request", req)

	// create response struct
	response := backend.NewQueryDataResponse()

	// loop over queries and execute them individually.
	for _, q := range req.Queries {
		res := d.query(ctx, req.PluginContext, q)

		// save the response in a hashmap
		// based on with RefID as identifier
		response.Responses[q.RefID] = res
	}

	return response, nil
}

type queryModel struct {
	QueryText string `json:"queryText"`
}

func (d *IOxDatasource) query(ctx context.Context, pCtx backend.PluginContext, query backend.DataQuery) backend.DataResponse {
	response := backend.DataResponse{}

	// Unmarshal the JSON into our queryModel.
	var qm queryModel

	response.Error = json.Unmarshal(query.JSON, &qm)
	if response.Error != nil {
		log.DefaultLogger.Error("unmarshal query failure", "query", string(query.JSON), "error", response.Error)
		return response
	}
	log.DefaultLogger.Debug("query called", "query", qm.QueryText)

	response.Frames, response.Error = d.query_client.query(ctx, d.Database, qm.QueryText)
	return response
}

// simplify a gRPC error to the underlying message, or return the provided error
// if the error didn't originate from gRPC.
func gRPCMsgFromErr(err error) error {
	s, ok := status.FromError(err)
	if ok {
		return errors.New(s.Message())
	}
	return err
}

func (d *IOxDatasource) connect() error {
	// Validate URL
	url, err := url.Parse(d.Host)
	if err != nil {
		return err
	}

	log.DefaultLogger.Debug("Dialling", "host", url.Host)
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)
	defer cancel()

	// TODO(edd): figure out correct options
	conn, err := grpc.DialContext(ctx, url.Host, grpc.WithInsecure(), grpc.WithBlock())
	if err != nil {
		log.DefaultLogger.Error("Connection error", "err", err)
		return err
	}
	d.mgt_client = pb.NewManagementServiceClient(conn)

	// Initialise a Flight client
	query_client, err := NewFlightClient(url.Host)
	if err != nil {
		return err
	}
	d.query_client = query_client
	return nil
}

// CheckHealth handles health checks sent from Grafana to the IOx backend.
// Health is checked by dialling a connection using the set host and ensuring
// that the specified database exists.
func (d *IOxDatasource) CheckHealth(ctx context.Context, req *backend.CheckHealthRequest) (*backend.CheckHealthResult, error) {
	log.DefaultLogger.Debug("CheckHealth called", "request", req)

	if d.err != nil { // underlying connection error
		err := gRPCMsgFromErr(d.err)
		return &backend.CheckHealthResult{
			Status:  backend.HealthStatusError,
			Message: fmt.Sprintf("Connection failed: %v", err),
		}, nil
	}

	in := pb.GetDatabaseRequest{Name: d.Database, OmitDefaults: true}
	if _, err := d.mgt_client.GetDatabase(context.Background(), &in); err != nil {
		log.DefaultLogger.Error("Unable to connect", "database", d.Database, "host", d.Host, "error", err)

		err = gRPCMsgFromErr(err)
		return &backend.CheckHealthResult{
			Status:  backend.HealthStatusError,
			Message: fmt.Sprintf("Connection failed: %v", err),
		}, nil
	}

	return &backend.CheckHealthResult{
		Status:  backend.HealthStatusOk,
		Message: "Data source is working",
	}, nil
}

// SubscribeStream is called when a client wants to connect to a stream. This callback
// allows sending the first message.
func (d *IOxDatasource) SubscribeStream(_ context.Context, req *backend.SubscribeStreamRequest) (*backend.SubscribeStreamResponse, error) {
	log.DefaultLogger.Debug("SubscribeStream called", "request", req)

	return &backend.SubscribeStreamResponse{
		Status: backend.SubscribeStreamStatusPermissionDenied,
	}, nil
}

// RunStream is called once for any open channel.  Results are shared with everyone
// subscribed to the same channel.
func (d *IOxDatasource) RunStream(ctx context.Context, req *backend.RunStreamRequest, sender *backend.StreamSender) error {
	log.DefaultLogger.Debug("RunStream called", "request", req)

	return nil
}

// PublishStream is called when a client sends a message to the stream.
func (d *IOxDatasource) PublishStream(_ context.Context, req *backend.PublishStreamRequest) (*backend.PublishStreamResponse, error) {
	log.DefaultLogger.Debug("PublishStream called", "request", req)

	// Do not allow publishing at all.
	return &backend.PublishStreamResponse{
		Status: backend.PublishStreamStatusPermissionDenied,
	}, nil
}
