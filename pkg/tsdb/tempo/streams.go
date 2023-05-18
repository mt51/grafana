package tempo

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/grafana/grafana-plugin-sdk-go/backend"
	"github.com/grafana/grafana-plugin-sdk-go/data"
	"github.com/grafana/grafana/pkg/infra/log"
	"github.com/grafana/tempo/pkg/tempopb"
	"sync"
)

type Streams struct {
	logger log.Logger

	// The running instances
	requests map[string]*TraceQLSearch

	// safe changes
	mutex sync.Mutex
}

const SearchPathPrefix = "search/"

func (s *Streams) add(key string, request *TraceQLSearch) error {
	if key == "" {
		return fmt.Errorf("invalid key %s", key)
	}

	s.mutex.Lock()
	defer s.mutex.Unlock()

	s.requests[key] = request
	return nil
}

func NewSearchStreams() *Streams {
	s := &Streams{
		requests: make(map[string]*TraceQLSearch),
		logger:   log.New("tsdb.tempo.search_streams"),
	}
	return s
}

func (s *Streams) runStream(ctx context.Context, req *backend.RunStreamRequest, sender *backend.StreamSender, tempoDatasource *TempoDatasource) error {
	s.logger.Info("Running stream", "path", req.Path, "req", req)

	sr, ok := s.requests[req.Path]
	// Ignore stream if the request was not found
	if !ok {
		return nil
	}

	if sr.State == Done {
		err := s.sendResponse(sr.Result, sender)
		if err != nil {
			return err
		}
	}

	select {
	case response := <-sr.Updates:
		err := s.sendResponse(response, sender)
		if err != nil {
			return err
		}
	case err := <-sr.Errors:
		return err
	}

	return nil
}

func (s *Streams) sendResponse(sr *tempopb.SearchResponse, sender *backend.StreamSender) error {
	frame := s.createResponseDataFrame()

	tracesAsJson, err := json.Marshal(sr.Traces)
	if err != nil {
		return err
	}
	tracesRawMessage := json.RawMessage(tracesAsJson)
	frame.Fields[0].Append(tracesRawMessage)

	metricsAsJson, err := json.Marshal(sr.Metrics)
	if err != nil {
		return err
	}
	metricsRawMessage := json.RawMessage(metricsAsJson)
	frame.Fields[1].Append(metricsRawMessage)

	err = sender.SendFrame(frame, data.IncludeAll)
	if err != nil {
		return err
	}

	return nil
}

func (s *Streams) createResponseDataFrame() *data.Frame {
	frame := data.NewFrame("response")
	frame.Fields = append(frame.Fields, data.NewField("traces", nil, []json.RawMessage{}))
	frame.Fields = append(frame.Fields, data.NewField("metrics", nil, []json.RawMessage{}))

	return frame
}

func (s *Streams) getRequestFromPath(path string) (*TraceQLSearch, error) {
	req, ok := s.requests[path]
	if !ok {
		return nil, fmt.Errorf("unable to find request for key %s", path)
	}
	return req, nil
}
