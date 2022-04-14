package segment

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"gopkg.in/segmentio/analytics-go.v3"
	"io/ioutil"
	"net/http"
	"plg-utilities/config"
	"time"
)

// The segment.HTTPClient is useful for workloads that require
// telemetry events to be sent synchronously such as CRON jobs.
// Use native go client segment.Client when you can,
// especially for web servers.

const (
	identityEndpoint = "/v1/identify"
	groupEndpoint    = "/v1/group"
	trackEndpoint    = "/v1/track"
	batchEndpoint    = "/v1/batch"

	batchSize = 20
)

type HTTPClient struct {
	enabled bool
	url     string
	apiKey  string
	client  *http.Client
}

// send identitfy, group,and track event
// send batch identitfy, group,and track event

func NewHTTPClient(config config.SegmentConf) *HTTPClient {
	return &HTTPClient{
		enabled: config.Enabled,
		url:     config.Url,
		apiKey:  config.ApiKey,
		client:  &http.Client{},
	}
}
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

	url := s.url + identityEndpoint
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

func (s *HTTPClient) SendBatchIdentifyEvent(identifyEvents []*analytics.Identify) error {
	if !s.enabled {
		return fmt.Errorf("skipping sending segment batch identify event")
	}
	if len(identifyEvents) > batchSize {
		return fmt.Errorf("batch size is greater than %d", batchSize)
	}

	for _, identifyEvent := range identifyEvents {
		if err := identifyEvent.Validate(); err != nil {
			return fmt.Errorf("could not validate batch identify event for %s: %s", identifyEvent.UserId, err.Error())
		}
	}

	b, err := json.Marshal(map[string]interface{}{"batch": identifyEvents})
	if err != nil {
		return fmt.Errorf("unable to marshal segment batch identify event: %s", err.Error())
	}

	url := s.url + batchEndpoint
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewReader(b))
	if err != nil {
		return fmt.Errorf("failed to create batch identify event: %s", err.Error())
	}

	req.Header.Add("Content-Type", "application/json")
	req.Header.Add("Content-Length", string(len(b)))
	req.SetBasicAuth(s.apiKey, "")

	res, err := s.client.Do(req)

	if err != nil {
		return fmt.Errorf("batch identify event request failed for user: %s", err.Error())
	}

	defer res.Body.Close()

	bt, err := ioutil.ReadAll(res.Body)
	fmt.Printf("doodle  %s\n", string(bt))
	return err
}

//
//func (s *HTTPClient) SendGroupEvent(accountId, identity string, traits map[string]interface{}, sendToAllDestinations bool,
//	destinations []string) error {
//	if !s.enabled {
//		logrus.Infof("skipping sending segment group event event for group %s by identity %s", accountId, identity)
//		return nil
//	}
//	analyticsTraits := buildAnalyticsTraits(traits)
//	integrations := buildIntegrations(sendToAllDestinations, destinations)
//	err := s.client.Enqueue(analytics.Group{UserId: identity, GroupId: accountId, Traits: analyticsTraits, Integrations: integrations})
//	if err != nil {
//		logrus.Errorf("unable to send segment group event for group %s: %s", accountId, err.Error())
//	}
//	return err
//}
//
//func (s *HTTPClient) SendTrackEvent(eventName, identity string, properties map[string]interface{}, sendToAllDestinations bool,
//	destinations []string) error {
//	if !s.enabled {
//		logrus.Infof("skipping sending segment track event %s by identity %s", eventName, identity)
//		return nil
//	}
//	analyticsProperties := buildAnalyticsProperties(properties)
//	integrations := buildIntegrations(sendToAllDestinations, destinations)
//	err := s.client.Enqueue(analytics.Track{UserId: identity, Properties: analyticsProperties, Integrations: integrations})
//	if err != nil {
//		logrus.Errorf("unable to send segment track event %s: %s", eventName, err.Error())
//	}
//	return err
//}
//
////func buildAnalyticsProperties(properties map[string]interface{}) analytics.Properties {
////	analyticsProperties := analytics.NewProperties()
////	for key, val := range properties {
////		if val == nil {
////			analyticsProperties.Set(key, "null")
////		} else {
////			analyticsProperties.Set(key, val)
////		}
////	}
////	return analyticsProperties
////}
////
////func buildIntegrations(sendToAllDestinations bool, destinations []string) analytics.Integrations {
////	integrations := analytics.Integrations{}
////	if sendToAllDestinations {
////		return integrations.EnableAll()
////	}
////	//todo: test if this is needed due to java behavior that sends to all destination by default
////	integrations.DisableAll()
////	for _, destination := range destinations {
////		integrations.Enable(destination)
////	}
////	return integrations
////}
////
////func buildAnalyticsTraits(traits map[string]interface{}) analytics.Traits {
////	analyticsTraits := analytics.NewTraits()
////	for key, val := range traits {
////		if val == nil {
////			analyticsTraits.Set(key, "null")
////		} else {
////			analyticsTraits.Set(key, val)
////		}
////	}
////	return analyticsTraits
////}

// Returns the time value passed as first argument, unless it's the zero-value,
// in that case the default value passed as second argument is returned.
func makeTimestamp(t time.Time, def time.Time) time.Time {
	if t == (time.Time{}) {
		return def
	}
	return t
}
