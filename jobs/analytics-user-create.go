package jobs

import (
	"context"
	"github.com/sirupsen/logrus"
	"gopkg.in/segmentio/analytics-go.v3"
	"plg-utilities/core"
	"plg-utilities/db/mongodb"
	"plg-utilities/telemetry/segment"
	"sync"
	"time"
)

func runAnalyticsUserCreate(mongo *mongodb.MongoDb, segmentSender *segment.HTTPClient) error {
	ctx := context.Background()
	var batchEvents []analytics.Message
	batchEventsQueue := make(chan []analytics.Message, 100)
	wg := sync.WaitGroup{}

	// three workers to send events
	for i := 0; i < 3; i++ {
		wg.Add(1)
		go func(segmentSender *segment.HTTPClient, wg *sync.WaitGroup) {
			defer wg.Done()
			for batchEvent := range batchEventsQueue {
				logrus.Infof("processing batch of events %+v", batchEvent)
				err := segmentSender.SendBatchEvents(batchEvent)
				if err != nil {
					logrus.Errorf("failed to process batch of events %+v", batchEvent)
				} else {
					logrus.Errorf("successful processing batch of events %+v", batchEvent)
				}
			}
		}(segmentSender, &wg)
	}

	// cache to track accountID to account Name
	// this is useful for user Group Event
	accountIdToName := map[string]string{}

	// process every account
	collectionName := mongo.AccountDAO.AccountCollection.Name()
	accountCursor, err := mongo.AccountDAO.ListWithCursor(ctx)
	if err != nil {
		logrus.Fatalf("unable to list collection %s: %s", collectionName, err.Error())
	}
	defer accountCursor.Close(ctx)
	for accountCursor.Next(ctx) {
		var account core.Account
		err := accountCursor.Decode(&account)
		if err != nil {
			logrus.Errorf("unable to decode record for collection %s with name %s: %s", collectionName, account.AccountName, err.Error())
		}
		// add to cache
		accountIdToName[account.Id] = account.AccountName
		createAccountIdentityEvent(account, &batchEvents, batchEventsQueue)
		createAccountGroupEvent(account, &batchEvents, batchEventsQueue)
	}

	if err := accountCursor.Err(); err != nil {
		logrus.Errorf("unable to list entire collection %s: %s", collectionName, err.Error())
	}

	// todo: users needed?
	// todo: send real user group event as part of Segment?
	// todo: also send real user identify event? this might overwrite existing user object's utm
	// possible chance that accountName could be null
	//process every user
	collectionName = mongo.UserDAO.UserCollections.Name()
	userCursor, err := mongo.UserDAO.ListWithCursor(ctx)
	if err != nil {
		logrus.Fatalf("unable to list collection %s: %s", collectionName, err.Error())
	}
	for userCursor.Next(ctx) {
		var user core.User
		err := userCursor.Decode(&user)
		if err != nil {
			logrus.Errorf("unable to decode record for collection %s with name %s: %s", collectionName, user.Name, err.Error())
		}
		createUserIdentityEvent(user, &batchEvents, batchEventsQueue)
		for _, accountId := range user.Accounts {
			createUserGroupEvent(user, accountId, accountIdToName[accountId], &batchEvents, batchEventsQueue)
		}
	}

	// flush batch events if there is any left
	if len(batchEvents) != 0 {
		flush(&batchEvents, batchEventsQueue)
	}

	if err := userCursor.Err(); err != nil {
		logrus.Errorf("unable to list entire collection %s: %s", collectionName, err.Error())
	}

	defer userCursor.Close(ctx)

	close(batchEventsQueue)
	wg.Wait()
	return err
}

func createAccountIdentityEvent(account core.Account, batchEvents *[]analytics.Message, queue chan []analytics.Message) {
	event := analytics.Identify{
		UserId:       segment.ACCOUNT_ANALYSIS_USER_PREFIX + account.Id,
		Timestamp:    time.Now(),
		Traits:       map[string]interface{}{"accountId": account.Id, "accountName": account.AccountName},
		Integrations: analytics.Integrations{}.DisableAll().Enable(segment.AMPLITUDE),
	}
	*batchEvents = append(*batchEvents, event)
	flushIfLimit(batchEvents, queue)
}

func createAccountGroupEvent(account core.Account, batchEvents *[]analytics.Message, queue chan []analytics.Message) {
	event := analytics.Group{
		UserId:       segment.ACCOUNT_ANALYSIS_USER_PREFIX + account.Id,
		GroupId:      account.Id,
		Timestamp:    time.Now(),
		Traits:       map[string]interface{}{"group_id": account.Id, "group_type": "Account", "group_name": account.AccountName},
		Integrations: analytics.Integrations{}.DisableAll().Enable(segment.AMPLITUDE),
	}
	*batchEvents = append(*batchEvents, event)
	flushIfLimit(batchEvents, queue)
}
func createUserIdentityEvent(user core.User, batchEvents *[]analytics.Message, queue chan []analytics.Message) {
	//properties.put("email", email);
	//properties.put("name", userInfo.getName());
	//properties.put("id", userInfo.getUuid());
	//properties.put("startTime", String.valueOf(Instant.now().toEpochMilli()));
	//properties.put("accountId", accountId);
	//properties.put("accountName", accountName);
	//properties.put("source", source);
	//properties.put("utm_source", utmInfo.getUtmSource() == null ? "" : utmInfo.getUtmSource());
	//properties.put("utm_content", utmInfo.getUtmContent() == null ? "" : utmInfo.getUtmContent());
	//properties.put("utm_medium", utmInfo.getUtmMedium() == null ? "" : utmInfo.getUtmMedium());
	//properties.put("utm_term", utmInfo.getUtmTerm() == null ? "" : utmInfo.getUtmTerm());
	//properties.put("utm_campaign", utmInfo.getUtmCampaign() == null ? "" : utmInfo.getUtmCampaign());
	event := analytics.Identify{
		UserId:    user.Email,
		Timestamp: time.Now(),
		Traits: map[string]interface{}{
			"email": user.Email,
			"name":  user.Name,
			"id":    user.Id,
			// not including accountId + accountName since they have multiple accounts
			//"accountId": user
			"source":       "migration",
			"utm_source":   user.UtmInfo.UtmSource,
			"utm_content":  user.UtmInfo.UtmSource,
			"utm_medium":   user.UtmInfo.UtmMedium,
			"utm_term":     user.UtmInfo.UtmTerm,
			"utm_campaign": user.UtmInfo.UtmCampaign,
		},
		Integrations: analytics.Integrations{}.DisableAll().Enable(segment.AMPLITUDE),
	}
	*batchEvents = append(*batchEvents, event)
	flushIfLimit(batchEvents, queue)
}

func createUserGroupEvent(user core.User, accountId string, accountName string,
	batchEvents *[]analytics.Message, queue chan []analytics.Message) {
	event := analytics.Group{
		UserId:       user.Email,
		GroupId:      accountId,
		Timestamp:    time.Now(),
		Traits:       map[string]interface{}{"group_id": accountId, "group_type": "Account", "group_name": accountName},
		Integrations: analytics.Integrations{}.DisableAll().Enable(segment.AMPLITUDE),
	}
	*batchEvents = append(*batchEvents, event)
	flushIfLimit(batchEvents, queue)

}

func flushIfLimit(batchEvents *[]analytics.Message, queue chan []analytics.Message) {
	if len(*batchEvents) > segment.BATCH_SIZE {
		flush(batchEvents, queue)
	}
}

func flush(batchEvents *[]analytics.Message, queue chan []analytics.Message) {
	queue <- *batchEvents
	*batchEvents = []analytics.Message{}
}

//func sendTelemetryEvents(account *core.Account, segmentSender *segment.HTTPClient) {
//	event := &analytics.Identify{
//		//Type: "",
//		//MessageId:    "",
//		//AnonymousId:  "",
//		UserId:    "test7",
//		Timestamp: time.Now(),
//		//Context:      nil,
//		Traits:       map[string]interface{}{"hi": "hi", "dodo": nil},
//		Integrations: nil,
//	}
//
//	//batch := []*analytics.Identify{
//	//	&analytics.Identify{
//	//		Type: "identify",
//	//		//MessageId:    "",
//	//		//AnonymousId:  "",
//	//		UserId:    "test8",
//	//		Timestamp: time.Now(),
//	//		//Context:      nil,
//	//		Traits:       map[string]interface{}{"hi": "hi", "dodo": nil},
//	//		Integrations: nil,
//	//	},
//	//	&analytics.Identify{
//	//		Type: "identify",
//	//		//MessageId:    "",
//	//		//AnonymousId:  "",
//	//		UserId:    "test9",
//	//		Timestamp: time.Now(),
//	//		//Context:      nil,
//	//		Traits:       map[string]interface{}{"hi": "hi", "dodo": nil},
//	//		Integrations: nil,
//	//	},
//	//}
//	//err := segmentSender.SendIdentifyEvent(&event)
//
//	batch := []analytics.Message{
//		analytics.Identify{
//			//Type: "identify",
//			//MessageId:    "",
//			//AnonymousId:  "",
//			UserId:    "test40",
//			Timestamp: time.Now(),
//			//Context:      nil,
//			Traits:       map[string]interface{}{"hi": "hi", "dodo": nil},
//			Integrations: nil,
//		},
//		analytics.Identify{
//			//Type: "identify",
//			//MessageId:    "",
//			//AnonymousId:  "",
//			UserId:    "test41",
//			Timestamp: time.Now(),
//			//Context:      nil,
//			Traits:       map[string]interface{}{"hi": "hi", "dodo": nil},
//			Integrations: nil,
//		},
//		&analytics.Identify{
//			//Type: "identify",
//			//MessageId:    "",
//			//AnonymousId:  "",
//			UserId:    "test42",
//			Timestamp: time.Now(),
//			//Context:      nil,
//			Traits:       map[string]interface{}{"hi": "hi", "dodo": nil},
//			Integrations: nil,
//		},
//		&analytics.Identify{
//			//Type: "identify",
//			//MessageId:    "",
//			//AnonymousId:  "",
//			UserId:    "test43",
//			Timestamp: time.Now(),
//			//Context:      nil,
//			Traits:       map[string]interface{}{"hi": "hi", "dodo": nil},
//			Integrations: nil,
//		},
//	}
//
//	err := segmentSender.SendBatchEvents(batch)
//	segmentSender.SendEvent(event)
//
//	//err := segmentSender.SendIdentifyEvent("test5", map[string]interface{}{"hi": "hi", "dodo": nil}, true, []string{})
//	if err != nil {
//		fmt.Printf("errorrrrrrr: %s\n", err.Error())
//	}
//}
