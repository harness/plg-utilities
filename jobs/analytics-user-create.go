package jobs

import (
	"context"
	"github.com/sirupsen/logrus"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo/options"
	"gopkg.in/segmentio/analytics-go.v3"
	"plg-utilities/core"
	"plg-utilities/db/mongodb"
	"plg-utilities/telemetry/segment"
	"sync"
	"time"
)

type MinimalAccount struct {
	Id          string `json:"_id,omitempty" bson:"_id,omitempty"`
	AccountName string `json:"accountName,omitempty" bson:"accountName,omitempty"`
}

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
					logrus.Errorf("failed to process batch of events %+v: %s", batchEvent, err.Error())
				} else {
					logrus.Infof("successful processing batch of events %+v", batchEvent)
				}
			}
		}(segmentSender, &wg)
	}

	// get all account Ids
	var accounts []MinimalAccount

	collectionName := mongo.AccountDAO.AccountCollection.Name()
	opts := options.Find().SetProjection(bson.D{{"_id", 1}, {"accountName", 1}})
	accountCursor, err := mongo.AccountDAO.AccountCollection.Find(ctx, bson.D{}, opts)
	if err != nil {
		logrus.Errorf("unable to find collection %s: %s", collectionName, err.Error())
	}
	if err := accountCursor.All(ctx, &accounts); err != nil {
		logrus.Errorf("unable to list collection %s: %s", collectionName, err.Error())
	}
	accountCursor.Close(ctx)

	// process every account and every use for that account
	collectionName = mongo.UserDAO.UserCollections.Name()
	for _, account := range accounts {
		// account telemetry
		createAccountIdentityEvent(account, &batchEvents, batchEventsQueue)
		createAccountGroupEvent(account, &batchEvents, batchEventsQueue)
		// user telemetry
		userCurser, err := mongo.UserDAO.UserCollections.Find(ctx, bson.M{"accounts": account.Id})
		if err != nil {
			logrus.Errorf("failed to find users for account %s: %s", account.Id, err.Error())
		}
		var users []core.User
		if err := userCurser.All(ctx, &users); err != nil {
			logrus.Errorf("failed to list users for account %s: %s", account.Id, err.Error())
		}
		userCurser.Close(ctx)
		for _, user := range users {
			createUserIdentityEvent(user, &batchEvents, batchEventsQueue)
			createUserGroupEvent(user, account.Id, account.AccountName, &batchEvents, batchEventsQueue)
		}
	}

	//// cache to track accountID to account Name
	//// this is useful for user Group Event
	//accountIdToName := map[string]string{}
	//
	//// process every account
	//collectionName := mongo.AccountDAO.AccountCollection.Name()
	//accountCursor, err := mongo.AccountDAO.ListWithCursor(ctx)
	//if err != nil {
	//	logrus.Fatalf("unable to list collection %s: %s", collectionName, err.Error())
	//}
	//defer accountCursor.Close(ctx)
	//for accountCursor.Next(ctx) {
	//	var account core.Account
	//	err := accountCursor.Decode(&account)
	//	if err != nil {
	//		logrus.Errorf("unable to decode record for collection %s with name %s: %s", collectionName, account.AccountName, err.Error())
	//	}
	//	// add to cache
	//	accountIdToName[account.Id] = account.AccountName
	//	createAccountIdentityEvent(account, &batchEvents, batchEventsQueue)
	//	createAccountGroupEvent(account, &batchEvents, batchEventsQueue)
	//}
	//
	//if err := accountCursor.Err(); err != nil {
	//	logrus.Errorf("unable to list entire collection %s: %s", collectionName, err.Error())
	//}
	//
	////process every user
	//collectionName = mongo.UserDAO.UserCollections.Name()
	//userCursor, err := mongo.UserDAO.ListWithCursor(ctx)
	//if err != nil {
	//	logrus.Fatalf("unable to list collection %s: %s", collectionName, err.Error())
	//}
	//for userCursor.Next(ctx) {
	//	var user core.User
	//	err := userCursor.Decode(&user)
	//	if err != nil {
	//		logrus.Errorf("unable to decode record for collection %s with name %s: %s", collectionName, user.Name, err.Error())
	//	}
	//	createUserIdentityEvent(user, &batchEvents, batchEventsQueue)
	//	for _, accountId := range user.Accounts {
	//		createUserGroupEvent(user, accountId, accountIdToName[accountId], &batchEvents, batchEventsQueue)
	//	}
	//}
	//
	//// flush batch events if there is any left
	//if len(batchEvents) != 0 {
	//	flush(&batchEvents, batchEventsQueue)
	//}
	//
	//if err := userCursor.Err(); err != nil {
	//	logrus.Errorf("unable to list entire collection %s: %s", collectionName, err.Error())
	//}
	//
	//defer userCursor.Close(ctx)

	close(batchEventsQueue)
	wg.Wait()
	return nil
}

func createAccountIdentityEvent(account MinimalAccount, batchEvents *[]analytics.Message, queue chan []analytics.Message) {
	event := analytics.Identify{
		UserId:       segment.ACCOUNT_ANALYSIS_USER_PREFIX + account.Id,
		Timestamp:    time.Now(),
		Traits:       map[string]interface{}{"accountId": account.Id, "accountName": account.AccountName},
		Integrations: analytics.Integrations{}.EnableAll(),
	}
	*batchEvents = append(*batchEvents, event)
	flushIfLimit(batchEvents, queue)
}

func createAccountGroupEvent(account MinimalAccount, batchEvents *[]analytics.Message, queue chan []analytics.Message) {
	event := analytics.Group{
		UserId:       segment.ACCOUNT_ANALYSIS_USER_PREFIX + account.Id,
		GroupId:      account.Id,
		Timestamp:    time.Now(),
		Traits:       map[string]interface{}{"group_id": account.Id, "group_type": "Account", "group_name": account.AccountName},
		Integrations: analytics.Integrations{}.EnableAll(),
	}
	*batchEvents = append(*batchEvents, event)
	flushIfLimit(batchEvents, queue)
}
func createUserIdentityEvent(user core.User, batchEvents *[]analytics.Message, queue chan []analytics.Message) {
	event := analytics.Identify{
		UserId:    user.Email,
		Timestamp: time.Now(),
		Traits: map[string]interface{}{
			"email":        user.Email,
			"name":         user.Name,
			"id":           user.Id,
			"source":       "migration",
			"utm_source":   user.UtmInfo.UtmSource,
			"utm_content":  user.UtmInfo.UtmSource,
			"utm_medium":   user.UtmInfo.UtmMedium,
			"utm_term":     user.UtmInfo.UtmTerm,
			"utm_campaign": user.UtmInfo.UtmCampaign,
		},
		Integrations: analytics.Integrations{}.EnableAll(),
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
		Integrations: analytics.Integrations{}.EnableAll(),
	}
	*batchEvents = append(*batchEvents, event)
	flushIfLimit(batchEvents, queue)

}

func flushIfLimit(batchEvents *[]analytics.Message, queue chan []analytics.Message) {
	if len(*batchEvents) == segment.BATCH_SIZE {
		flush(batchEvents, queue)
	}
}

func flush(batchEvents *[]analytics.Message, queue chan []analytics.Message) {
	queue <- *batchEvents
	*batchEvents = []analytics.Message{}
}
