package jobs

import (
	"plg-utilities/core"
	"plg-utilities/db/mongodb"
	"plg-utilities/telemetry/segment"
)

//todo: add better logs
//todo: pipeline for job using kubernetes
//todo: env variables override in diff envs
//todo: check if events being sent that are not shown in segment and show in amplitude
func runAnalyticsUserCreate(mongo *mongodb.MongoDb, segmentSender *segment.Segment) error {
	defer segmentSender.Close()
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

func sendTelemetryEvents(account *core.Account, segmentSender *segment.Segment) {
	segmentSender.SendIdentifyEvent("test", map[string]interface{}{"hi": "hi"}, true, []string{})
}
