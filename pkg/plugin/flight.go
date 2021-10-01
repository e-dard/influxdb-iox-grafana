package plugin

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/apache/arrow/go/arrow"
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

		// TODO(edd):
		//
		// Figure out how to determine whether to emit long or wide data format.
		//
		// I expected to be able to handle this on the front-end via
		// transformations but it doesn't seem to be working for me.
		hasTimeColumn := false
		hasLabelColumn := false
		for _, field := range rb.Schema().Fields() {
			if field.Name == "time" && field.Type.ID() != arrow.STRING {
				hasTimeColumn = true
			} else if field.Type.ID() == arrow.STRING {
				hasLabelColumn = true
			}
		}

		frame, err := data.FromArrowRecord(rb)
		if err != nil {
			log.DefaultLogger.Error("Unable to convert RB to Frame", "error", err)
			return nil, err
		}

		if hasLabelColumn && hasTimeColumn {
			log.DefaultLogger.Debug("Assuming time series WIDE format", "columns", fmt.Sprintf("%v", rb.Columns()))
			if frame, err = data.LongToWide(frame, nil); err != nil {
				log.DefaultLogger.Error("Unable to convert frame from long to wide", "error", err)
				return nil, err
			}
		}

		frames = append(frames, frame)
	}

	return frames, nil
}
