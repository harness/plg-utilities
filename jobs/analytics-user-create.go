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
	collectionName := mongo.AccountDAO.AccountCollection.Name()
	cursor, err := mongo.AccountDAO.ListWithCursor(ctx)
	if err != nil {
		logrus.Fatalf("unable to list collection %s: %s", collectionName, err.Error())
	}
	defer cursor.Close(ctx)

	batchEventQueue := make(chan []analytics.Message, 100)
	wg := sync.WaitGroup{}

	// three workers to send events
	for i := 0; i < 3; i++ {
		wg.Add(1)
		go func(segmentSender *segment.HTTPClient, wg *sync.WaitGroup) {
			defer wg.Done()
			for batchEvent := range batchEventQueue {
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

	var batchEvent []analytics.Message
	for cursor.Next(ctx) {
		var account core.Account
		err := cursor.Decode(&account)
		if err != nil {
			logrus.Errorf("unable to decode record for collection %s with name%s: %s", collectionName, account.AccountName, err.Error())
		}
		if len(batchEvent) > segment.BATCH_SIZE {
			batchEventQueue <- batchEvent
			batchEvent = []analytics.Message{}
		}
		identityEvent := createIdentityEvent(account)
		groupEvent := createGroupEvent(account)
		batchEvent = append(batchEvent, identityEvent, groupEvent)
	}
	if len(batchEvent) <= segment.BATCH_SIZE {
		batchEventQueue <- batchEvent
	}

	if err := cursor.Err(); err != nil {
		logrus.Errorf("unable to list entire collection %s: %s", collectionName, err.Error())
	}

	close(batchEventQueue)
	wg.Wait()
	return err
}

func createIdentityEvent(account core.Account) analytics.Message {
	return analytics.Identify{
		UserId:       segment.ACCOUNT_ANALYSIS_USER_PREFIX + account.Id,
		Timestamp:    time.Now(),
		Traits:       map[string]interface{}{"accountId": account.Id, "accountName": account.AccountName},
		Integrations: analytics.Integrations{}.DisableAll().Enable(segment.AMPLITUDE),
	}
}

func createGroupEvent(account core.Account) analytics.Message {
	return analytics.Group{
		UserId:       segment.ACCOUNT_ANALYSIS_USER_PREFIX + account.Id,
		GroupId:      account.Id,
		Timestamp:    time.Now(),
		Traits:       map[string]interface{}{"group_id": account.Id, "group_type": "Account", "group_name": account.AccountName},
		Integrations: analytics.Integrations{}.DisableAll().Enable(segment.AMPLITUDE),
	}
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
