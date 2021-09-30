package plugin

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/apache/arrow/go/arrow/flight"
	"github.com/grafana/grafana-plugin-sdk-go/backend/log"
	"github.com/grafana/grafana-plugin-sdk-go/data"
	"google.golang.org/grpc"
)

type FlightClient struct {
	client flight.Client
}

func NewFlightClient(addr string) (*FlightClient, error) {
	client, err := flight.NewFlightClient(addr, nil, grpc.WithInsecure(), grpc.WithBlock())
	if err != nil {
		return nil, err
	}
	return &FlightClient{client: client}, nil
}

type IOxFlightQuery struct {
	DatabaseName string `json:"database_name"`
	Query        string `json:"sql_query"`
}

func (c *FlightClient) query(ctx context.Context, database string, query string) (data.Frames, error) {
	log.DefaultLogger.Debug("flight query called", "database", database, "query", query)

	//marshal payload into JSON
	payload, err := json.Marshal(IOxFlightQuery{
		DatabaseName: database,
		Query:        query,
	})
	if err != nil {
		return nil, err
	}

	resp, err := c.client.DoGet(ctx, &flight.Ticket{Ticket: payload})
	if err != nil {
		log.DefaultLogger.Error("flight DoGet error", "error", err)
		return nil, err
	}

	r, err := flight.NewRecordReader(resp)
	if err != nil {
		return nil, err
	}
	defer r.Release()

	var frames data.Frames
	for r.Next() {
		r.Schema()
		rb := r.Record()
		defer rb.Release()
		log.DefaultLogger.Debug("Record Batch contents", "columns", fmt.Sprintf("%v", rb.Columns()))

		frame, err := data.FromArrowRecord(rb)
		if err != nil {
			log.DefaultLogger.Error("Unable to convert RB to Frame", "error", err)
			return nil, err
		}
		frames = append(frames, frame)
	}

	return frames, nil
}
