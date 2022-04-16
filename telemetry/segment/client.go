package segment

//
//const (
//	MARKETO    = "Marketo"
//	SALESFORCE = "Salesforce"
//	AMPLITUDE  = "Amplitude"
//)
//
//type Client struct {
//	// checks if segment is enabled
//	enabled bool
//	// segment analytics client
//	client analytics.Client
//	// used to block and log for segment events
//	msgTracker *segmentMsgTracker
//}
//
//
//// NewClient creates a Segment client
//// logger is an optional parameter to log event processing using callbacks
//func NewClient(config config.SegmentConf, logger Logger) (*Client, error) {
//	analyticsConfig := analytics.Config{
//		Logger:  logger,
//		Verbose: true,
//		// logs event processing + success + failures
//		Callback: &segmentMsgTracker{
//			logger: logger,
//		},
//	}
//
//	client, err := analytics.NewWithConfig(config.ApiKey, analyticsConfig)
//	if err != nil {
//		return nil, err
//	}
//
//	return &Client{
//		enabled: config.Enabled,
//		client:  client,
//		msgTracker: &segmentMsgTracker{
//			logger: logger,
//		},
//	}, nil
//}
//
//func (s *Client) SendIdentifyEvent(identity string, traits map[string]interface{}, sendToAllDestinations bool,
//	destinations []string) error {
//	if !s.enabled {
//		return fmt.Errorf("skipping sending segment identify event for %s", identity)
//	}
//	analyticsTraits := buildAnalyticsTraits(traits)
//	integrations := buildIntegrations(sendToAllDestinations, destinations)
//	identityEvent := analytics.Identify{UserId: identity, Traits: analyticsTraits, Integrations: integrations}
//	err := s.client.Enqueue(identityEvent)
//	if err != nil {
//		return fmt.Errorf("unable to send segment identify event for identity %s: %s", identity, err.Error())
//
//	}
//	s.msgTracker.Add(identityEvent)
//	return err
//}
//
//func (s *Client) SendGroupEvent(accountId, identity string, traits map[string]interface{}, sendToAllDestinations bool,
//	destinations []string) error {
//	if !s.enabled {
//		return fmt.Errorf("skipping sending segment group event event for group %s by identity %s", accountId, identity)
//	}
//	analyticsTraits := buildAnalyticsTraits(traits)
//	integrations := buildIntegrations(sendToAllDestinations, destinations)
//	groupEvent := analytics.Group{UserId: identity, GroupId: accountId, Traits: analyticsTraits, Integrations: integrations}
//	err := s.client.Enqueue(groupEvent)
//	if err != nil {
//		return fmt.Errorf("unable to send segment group event for group %s: %s", accountId, err.Error())
//	}
//	s.msgTracker.Add(groupEvent)
//	return err
//}
//
//func (s *Client) SendTrackEvent(eventName, identity string, properties map[string]interface{}, sendToAllDestinations bool,
//	destinations []string) error {
//	if !s.enabled {
//		return fmt.Errorf("skipping sending segment track event %s by identity %s", eventName, identity)
//	}
//	analyticsProperties := buildAnalyticsProperties(properties)
//	integrations := buildIntegrations(sendToAllDestinations, destinations)
//	trackEvent := analytics.Track{UserId: identity, Properties: analyticsProperties, Integrations: integrations}
//	err := s.client.Enqueue(trackEvent)
//	if err != nil {
//		return fmt.Errorf("unable to send segment track event %s: %s", eventName, err.Error())
//	}
//	s.msgTracker.Add(trackEvent)
//	return err
//}
//
//func (s *Client) Close() error {
//	return s.client.Close()
//}
//
//func buildAnalyticsProperties(properties map[string]interface{}) analytics.Properties {
//	analyticsProperties := analytics.NewProperties()
//	for key, val := range properties {
//		if val == nil {
//			analyticsProperties.Set(key, "null")
//		} else {
//			analyticsProperties.Set(key, val)
//		}
//	}
//	return analyticsProperties
//}
//
//func buildIntegrations(sendToAllDestinations bool, destinations []string) analytics.Integrations {
//	integrations := analytics.Integrations{}
//	if sendToAllDestinations {
//		return integrations.EnableAll()
//	}
//	//todo: test if this is needed due to java sdk behavior that sends to all destination by default
//	integrations.DisableAll()
//	for _, destination := range destinations {
//		integrations.Enable(destination)
//	}
//	return integrations
//}
//
//func buildAnalyticsTraits(traits map[string]interface{}) analytics.Traits {
//	analyticsTraits := analytics.NewTraits()
//	for key, val := range traits {
//		analyticsTraits.Set(key, val)
//		//if val == nil {
//		//	analyticsTraits.Set(key, "null")
//		//} else {
//		//	analyticsTraits.Set(key, val)
//		//}
//	}
//	return analyticsTraits
//}
//
//// used to log successful and failed message sends to Segment
//type segmentMsgTracker struct {
//	logger Logger
//}
//
//// Add logs that messages are being processed
//func (s *segmentMsgTracker) Add(message analytics.Message) {
//	if s.logger != nil {
//		s.logger.Logf("processing message: %v", message)
//	}
//}
//
//// Success logs that messages are successfully processed via callback
//func (s *segmentMsgTracker) Success(message analytics.Message) {
//	if s.logger != nil {
//		s.logger.Logf("successfully sent message: %v", message)
//	}
//}
//
//// Failure logs that messages that failed to process via callback
//func (s *segmentMsgTracker) Failure(message analytics.Message, err error) {
//	if s.logger != nil {
//		s.logger.Errorf("failed to send message: %v", message)
//	}
//}
