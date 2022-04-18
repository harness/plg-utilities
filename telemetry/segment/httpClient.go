package segment

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"gopkg.in/segmentio/analytics-go.v3"
	"net/http"
	"plg-utilities/config"
	"time"
)

// The segment.HTTPClient is useful for workloads that require
// telemetry events to be sent synchronously such as CRON jobs.
// Use native go client segment.Client when you can,
// especially for web servers that require auto-batching.

// todo: send account identify + group event to Segment - done
// todo: check if events being sent that are not shown in segment and show in amplitude
// todo: add unit tests
// todo: create github + webhook triggers - done
// todo: create build pipeline - in progress
// todo: override env variables in diff envs
// todo: create deployment pipeline
type HTTPClient struct {
	enabled bool
	url     string
	apiKey  string
	client  *http.Client
}

func NewHTTPClient(config config.SegmentConf) *HTTPClient {
	return &HTTPClient{
		enabled: config.Enabled,
		url:     config.Url,
		apiKey:  config.ApiKey,
		client:  &http.Client{},
	}
}

func (s *HTTPClient) SendEvent(event analytics.Message) error {
	if !s.enabled {
		return fmt.Errorf("skipping sending segment event %+v", event)
	}
	if err := event.Validate(); err != nil {
		return fmt.Errorf("could not validate event %s: %s", event, err.Error())
	}

	event, msgType, err := formatMessage(event)
	if err != nil {
		return err
	}

	// get the correct url for event type
	url := s.url
	switch msgType {
	case GROUP:
		url += GROUP_ENDPOINT
	case IDENTIFY:
		url += IDENTIFY_ENDPOINT
	case TRACK:
		url += TRACK_ENDPOINT
	default:
		return fmt.Errorf("incorrect event type %s for msg %+v", msgType, event)
	}

	b, err := json.Marshal(event)
	if err != nil {
		return fmt.Errorf("unable to marshal segment event %s: %s", event, err.Error())
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewReader(b))
	if err != nil {
		return fmt.Errorf("failed to create event request %s: %s", event, err.Error())
	}

	req.Header.Add("Content-Type", "application/json")
	req.SetBasicAuth(s.apiKey, "")

	_, err = s.client.Do(req)

	if err != nil {
		return fmt.Errorf("identify event request failed %s: %s", event, err.Error())
	}
	return err
}

type batch struct {
	SentAt   time.Time           `json:"sentAt"`
	Messages []analytics.Message `json:"batch"`
}

func (s *HTTPClient) SendBatchEvents(messages []analytics.Message) error {
	if !s.enabled {
		return fmt.Errorf("skipping sending segment batch events")
	}
	if len(messages) > BATCH_SIZE {
		return fmt.Errorf("batch size is greater than limit of %d", BATCH_SIZE)
	}

	// modify messages to be in the correct format for Segment
	for i, msg := range messages {
		if err := msg.Validate(); err != nil {
			return fmt.Errorf("could not validate batch event %+v: %s", msg, err.Error())
		}
		m, _, err := formatMessage(msg)
		if err != nil {
			return err
		}
		messages[i] = m
	}

	b, err := json.Marshal(batch{Messages: messages, SentAt: time.Now()})

	if err != nil {
		return fmt.Errorf("unable to marshal segment batch event: %s", err.Error())
	}

	url := s.url + BATCH_ENDPOINT
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewReader(b))
	if err != nil {
		return fmt.Errorf("failed to create batch event: %s", err.Error())
	}

	req.Header.Add("Content-Type", "application/json")
	req.SetBasicAuth(s.apiKey, "")

	_, err = s.client.Do(req)
	if err != nil {
		return fmt.Errorf("batch identify event request failed: %s", err.Error())
	}
	return err
}

func formatMessage(msg analytics.Message) (analytics.Message, string, error) {
	ts := time.Now()
	switch m := msg.(type) {
	case *analytics.Group:
		m.Type = GROUP
		m.Timestamp = makeTimestamp(m.Timestamp, ts)
		return m, GROUP, nil
	case *analytics.Identify:
		m.Type = IDENTIFY
		m.Timestamp = makeTimestamp(m.Timestamp, ts)
		return m, IDENTIFY, nil
	case *analytics.Track:
		m.Type = TRACK
		m.Timestamp = makeTimestamp(m.Timestamp, ts)
		return m, TRACK, nil
	case analytics.Group:
		m.Type = GROUP
		m.Timestamp = makeTimestamp(m.Timestamp, ts)
		return m, GROUP, nil
	case analytics.Identify:
		m.Type = IDENTIFY
		m.Timestamp = makeTimestamp(m.Timestamp, ts)
		return m, IDENTIFY, nil
	case analytics.Track:
		m.Type = TRACK
		m.Timestamp = makeTimestamp(m.Timestamp, ts)
		return m, TRACK, nil
	default:
		return nil, "", fmt.Errorf("incorrect type %T message for %+v", m, msg)
	}
}

// Returns the time value passed as first argument, unless it's the zero-value,
// in that case the default value passed as second argument is returned.
func makeTimestamp(t time.Time, def time.Time) time.Time {
	if t == (time.Time{}) {
		return def
	}
	return t
}

// SendIdentifyEvent - redundant
// todo: possibly add more send event methods but
// currently this is unnecessary since SendEvent can be used to any event type
func (s *HTTPClient) SendIdentifyEvent(identifyEvent *analytics.Identify) error {
	if !s.enabled {
		return fmt.Errorf("skipping sending segment identify event for %s", identifyEvent.UserId)
	}
	if err := identifyEvent.Validate(); err != nil {
		return fmt.Errorf("could not validate identify event for %s: %s", identifyEvent.UserId, err.Error())
	}

	b, err := json.Marshal(identifyEvent)
	if err != nil {
		return fmt.Errorf("unable to marshal segment identify event for user %s: %s", identifyEvent.UserId, err.Error())
	}

	url := s.url + IDENTIFY_ENDPOINT
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewReader(b))
	if err != nil {
		return fmt.Errorf("failed to create identify event request for user %s: %s", identifyEvent.UserId, err.Error())
	}

	req.Header.Add("Content-Type", "application/json")
	req.Header.Add("Content-Length", string(len(b)))
	req.SetBasicAuth(s.apiKey, "")

	res, err := s.client.Do(req)

	if err != nil {
		return fmt.Errorf("identify event request failed for user %s: %s", identifyEvent.UserId, err.Error())
	}

	defer res.Body.Close()
	return err
}
