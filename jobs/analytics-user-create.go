package jobs

import (
	"fmt"
	"gopkg.in/segmentio/analytics-go.v3"
	"plg-utilities/core"
	"plg-utilities/db/mongodb"
	"plg-utilities/telemetry/segment"
	"time"
)

//todo: add better logs
//todo: pipeline for job using kubernetes
//todo: env variables override in diff envs
//todo: check if events being sent that are not shown in segment and show in amplitude
func runAnalyticsUserCreate(mongo *mongodb.MongoDb, segmentSender *segment.HTTPClient) error {
	//defer segmentSender.Close()
	sendTelemetryEvents(&core.Account{}, segmentSender)
	return nil
	//ctx, _ := context.WithTimeout(context.Background(), 10*time.Second)
	//collectionName := mongo.AccountDAO.AccountCollection.Name()
	//cursor, err := mongo.AccountDAO.ListWithCursor(ctx)
	//if err != nil {
	//	logrus.Fatalf("unable to list collection %s: %s", collectionName, err.Error())
	//}
	//defer cursor.Close(ctx)
	//
	////var accounts []core.Account
	//for cursor.Next(ctx) {
	//	var account core.Account
	//	err := cursor.Decode(&account)
	//	if err != nil {
	//		logrus.Errorf("unable to decode record for collection %s: %s", collectionName, err.Error())
	//	}
	//	//todo: remove log statement
	//	sendTelemetryEvents(&account, segmentSender)
	//	log.Printf("%v\n", account)
	//	//accounts = append(accounts, account)
	//}
	//
	//if err := cursor.Err(); err != nil {
	//	logrus.Errorf("unable to list entire collection %s: %s", collectionName, err.Error())
	//}
	//return err
}

func sendTelemetryEvents(account *core.Account, segmentSender *segment.HTTPClient) {
	//event := analytics.Identify{
	//	//Type: "",
	//	//MessageId:    "",
	//	//AnonymousId:  "",
	//	UserId:    "test7",
	//	Timestamp: time.Now(),
	//	//Context:      nil,
	//	Traits:       map[string]interface{}{"hi": "hi", "dodo": nil},
	//	Integrations: nil,
	//}

	batch := []*analytics.Identify{
		&analytics.Identify{
			Type: "identify",
			//MessageId:    "",
			//AnonymousId:  "",
			UserId:    "test8",
			Timestamp: time.Now(),
			//Context:      nil,
			Traits:       map[string]interface{}{"hi": "hi", "dodo": nil},
			Integrations: nil,
		},
		&analytics.Identify{
			Type: "identify",
			//MessageId:    "",
			//AnonymousId:  "",
			UserId:    "test9",
			Timestamp: time.Now(),
			//Context:      nil,
			Traits:       map[string]interface{}{"hi": "hi", "dodo": nil},
			Integrations: nil,
		},
	}
	//err := segmentSender.SendIdentifyEvent(&event)

	err := segmentSender.SendBatchIdentifyEvent(batch)

	//err := segmentSender.SendIdentifyEvent("test5", map[string]interface{}{"hi": "hi", "dodo": nil}, true, []string{})
	if err != nil {
		fmt.Printf("errorrrrrrr: %s\n", err.Error())
	}
}
