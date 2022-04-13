package segment

import (
	"fmt"
	"github.com/sirupsen/logrus"
	"gopkg.in/segmentio/analytics-go.v3"
	"plg-utilities/config"
)

const (
	MARKETO    = "Marketo"
	SALESFORCE = "Salesforce"
	AMPLITUDE  = "Amplitude"
)

type Segment struct {
	// checks if segment is enabled
	enabled bool
	// segment analytics client
	client analytics.Client
	// used to block and log for segment events
	msgTracker *segmentMsgTracker

	block chan bool
}

// todo: make logger a constructor param and make it used by Logger and Callback func to log success + failures - done
// todo: do something about event logs like track event skip logs
// todo: send account identify + group event to Segment
// todo: figure out which db to get account details from - it should be CG - done
// todo: send user group event as part of Semgnet. Should I also send user identify event? this might overwrite existing user object's utm
// todo: add unit tests
// todo: create github + webhook triggers
// todo: create deployment pipeline
//todo: make sure code to synchronize events is good
func New(config config.SegmentConf, logger Logger) (*Segment, error) {
	if logger == nil {
		logger = &DefaultSegmentLogger{}
	}

	// set up message tracker if segment messages
	// should block until all events are sent to Segment
	msgTracker := segmentMsgTracker{
		logger: logger,
		block:  false,
	}

	analyticsConfig := analytics.Config{
		Logger:  logger,
		Verbose: true,
		// logging process + success + failures
		Callback:  &msgTracker,
		BatchSize: 1,
	}

	if config.BlockForEvents {
		//todo: make this not 10
		msgTracker.MsgTracker = make(chan bool, 10)
		msgTracker.block = true
		analyticsConfig.Callback = &msgTracker
	}

	client, err := analytics.NewWithConfig(config.ApiKey, analyticsConfig)
	if err != nil {
		logrus.Errorf("unable to connect to segment: %s", err.Error())
		return nil, err
	}

	return &Segment{
		enabled:    config.Enabled,
		client:     client,
		msgTracker: &msgTracker,
		block:      make(chan bool),
	}, nil

}

func (s *Segment) SendIdentifyEvent(identity string, traits map[string]interface{}, sendToAllDestinations bool,
	destinations []string) error {
	for i := 0; i < 100; i++ {
		msg := analytics.Identify{
			UserId: "019mr8mf4r",
			Traits: analytics.NewTraits().
				SetName(fmt.Sprintf("Michael Bolton2 %d", i)).
				SetEmail("mbolton@example.com").
				Set("plan", "Enterprise").
				Set("friends", 42),
		}
		s.msgTracker.Add(msg)
		s.client.Enqueue(msg)
	}

	//<-s.MsgTracker.MsgTracker
	//close(s.MsgTracker.MsgTracker)
	return nil
	//if !s.enabled {
	//	logrus.Infof("skipping sending segment identify event for %s", identity)
	//	return nil
	//}
	//analyticsTraits := buildAnalyticsTraits(traits)
	//integrations := buildIntegrations(sendToAllDestinations, destinations)
	//err := s.client.Enqueue(analytics.Identify{UserId: identity, Traits: analyticsTraits, Integrations: integrations})
	//if err != nil {
	//	logrus.Infof("unable to send segment identify event for identity %s: %s", identity, err.Error())
	//
	//	logrus.Errorf("unable to send segment identify event for identity %s: %s", identity, err.Error())
	//}
	//return err
}

func (s *Segment) SendGroupEvent(accountId, identity string, traits map[string]interface{}, sendToAllDestinations bool,
	destinations []string) error {
	if !s.enabled {
		logrus.Infof("skipping sending segment group event event for group %s by identity %s", accountId, identity)
		return nil
	}
	analyticsTraits := buildAnalyticsTraits(traits)
	integrations := buildIntegrations(sendToAllDestinations, destinations)
	err := s.client.Enqueue(analytics.Group{UserId: identity, GroupId: accountId, Traits: analyticsTraits, Integrations: integrations})
	if err != nil {
		logrus.Errorf("unable to send segment group event for group %s: %s", accountId, err.Error())
	}
	return err
}

func (s *Segment) SendTrackEvent(eventName, identity string, properties map[string]interface{}, sendToAllDestinations bool,
	destinations []string) error {
	if !s.enabled {
		logrus.Infof("skipping sending segment track event %s by identity %s", eventName, identity)
		return nil
	}
	analyticsProperties := buildAnalyticsProperties(properties)
	integrations := buildIntegrations(sendToAllDestinations, destinations)
	err := s.client.Enqueue(analytics.Track{UserId: identity, Properties: analyticsProperties, Integrations: integrations})
	if err != nil {
		logrus.Errorf("unable to send segment track event %s: %s", eventName, err.Error())
	}
	return err
}

func (s *Segment) Close() error {
	//return s.client.Close()
	for {
		fmt.Printf("ASDSADASFASDFSAFGSADFDSA\n")
		if s.msgTracker.processed == s.msgTracker.queued {
			return nil
		}
	}
}

func buildAnalyticsProperties(properties map[string]interface{}) analytics.Properties {
	analyticsProperties := analytics.NewProperties()
	for key, val := range properties {
		if val == nil {
			analyticsProperties.Set(key, "null")
		} else {
			analyticsProperties.Set(key, val)
		}
	}
	return analyticsProperties
}

func buildIntegrations(sendToAllDestinations bool, destinations []string) analytics.Integrations {
	integrations := analytics.Integrations{}
	if sendToAllDestinations {
		return integrations.EnableAll()
	}
	//todo: test if this is needed due to java behavior that sends to all destination by default
	integrations.DisableAll()
	for _, destination := range destinations {
		integrations.Enable(destination)
	}
	return integrations
}

func buildAnalyticsTraits(traits map[string]interface{}) analytics.Traits {
	analyticsTraits := analytics.NewTraits()
	for key, val := range traits {
		if val == nil {
			analyticsTraits.Set(key, "null")
		} else {
			analyticsTraits.Set(key, val)
		}
	}
	return analyticsTraits
}

type segmentMsgTracker struct {
	// checks if current goroutine should be blocked
	// if there are pending messages to be sent to Segment
	block bool
	// channel to keep track of messages and block current goroutine
	MsgTracker chan bool
	// used to log successfully and failed message sends to Segment
	logger Logger

	queued    int
	processed int
}

// Add adds segment message to channel so that current goroutine will block
// until message is read from msg channel
func (s *segmentMsgTracker) Add(message analytics.Message) {
	s.queued++
	fmt.Printf("hi add %d %d %d\n", len(s.MsgTracker), s.queued, s.processed)
	//if s.block {
	//	s.MsgTracker <- true
	//}
	//s.logger.Logf("processing message: %v", message)
}

// Success blocks current goroutine until a message arrives to msg channel
// and logs successful messages send to segment
func (s *segmentMsgTracker) Success(message analytics.Message) {
	s.processed++
	fmt.Printf("hi success %d %d %d\n", len(s.MsgTracker), s.queued, s.processed)
	//if s.block {
	//	<-s.MsgTracker
	//}
	//s.logger.Logf("successfully sent message: %v", message)
}

// Failure blocks current goroutine until a message arrives to msg channel
// and logs failures
func (s *segmentMsgTracker) Failure(message analytics.Message, err error) {
	s.processed++
	fmt.Printf("hi fail%d %d %d\n", len(s.MsgTracker), s.queued, s.processed)
	//if s.block {
	//	<-s.MsgTracker
	//}
	//s.logger.Errorf("failed to send message: %v", message)
}
