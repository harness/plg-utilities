package cronjobs

import (
	"context"
	"fmt"
	"plg-utilities/core"
	"plg-utilities/db/mongodb"
	"plg-utilities/telemetry/segment"
	"strconv"
	"sync"
	"time"

	"github.com/sirupsen/logrus"
	"go.mongodb.org/mongo-driver/bson"
	"gopkg.in/segmentio/analytics-go.v3"
)

const (
	ISOformat      = "20060102"
	ISOMonthFormat = "200601"
)

func RunAccountTraitsJob(mongo *mongodb.MongoDb, segmentSender *segment.HTTPClient) error {
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

	// process every account
	collectionName := mongo.AccountDAO.AccountCollection.Name()
	accountsCursor, err := mongo.AccountDAO.ListWithCursor(ctx)
	if err != nil {
		logrus.Fatalf("unable to list collection %s: %s", collectionName, err.Error())
	}

	defer accountsCursor.Close(ctx)
	for accountsCursor.Next(ctx) {
		var account core.Account
		err := accountsCursor.Decode(&account)
		if err != nil {
			logrus.Errorf("unable to decode record for collection %s: %+v: %s", collectionName, account, err.Error())
			continue
		}

		//get moduleLicenses for the Account
		moduleLicenseCursor, err := mongo.ModuleLicenseDAO.ModuleLicenseCollection.Find(ctx, bson.M{"accountIdentifier": account.Id})
		if err != nil {
			logrus.Fatalf("unable to list collection %s: %s", collectionName, err.Error())
			continue
		}
		var moduleLicenses []core.ModuleLicense
		if err := moduleLicenseCursor.All(ctx, &moduleLicenses); err != nil {
			logrus.Errorf("failed to get all moduleLicenses for account %s: %s", account.Id, err.Error())
			continue
		}

		logrus.Infof("found for account %s modules licenses +%v", account.Id, moduleLicenses)

		//err = moduleLicenseCursor.Decode(&moduleLicenses)
		//if err != nil {
		//	logrus.Errorf("unable to decode record for collection %s: %s", collectionName, err.Error())
		//	continue
		//}

		logrus.Infof("found in collection %s: %+v", collectionName, account)
		createAccountGroupEvent(account, moduleLicenses, &batchEvents, batchEventsQueue)
	}

	// flush batch events if there is any left
	if len(batchEvents) != 0 {
		flush(&batchEvents, batchEventsQueue)
	}

	if err := accountsCursor.Err(); err != nil {
		logrus.Errorf("unable to list entire collection %s: %s", collectionName, err.Error())
	}

	close(batchEventsQueue)
	wg.Wait()
	return nil
}

func createAccountGroupEvent(account core.Account, moduleLicenses []core.ModuleLicense, batchEvents *[]analytics.Message, queue chan []analytics.Message) {
	accountId := account.Id
	isPaid := isAccountPaid(moduleLicenses)
	created_at, created_at_week, created_at_month := createdAtInfo(account.CreatedAt)
	traits := map[string]interface{}{"group_id": accountId, "group_type": "Account", "is_paid": isPaid, "created_at": created_at, "created_at_week": created_at_week, "created_at_month": created_at_month}

	event := analytics.Group{
		UserId:       segment.ACCOUNT_ANALYSIS_USER_PREFIX + accountId,
		GroupId:      accountId,
		Timestamp:    time.Now(),
		Traits:       traits,
		Integrations: analytics.Integrations{}.EnableAll(),
	}
	*batchEvents = append(*batchEvents, event)
	flushIfLimit(batchEvents, queue)
}

func isAccountPaid(moduleLicenses []core.ModuleLicense) bool {
	for _, moduleLicense := range moduleLicenses {
		if moduleLicense.LicenseType == "PAID" {
			return true
		}
	}
	return false
}

func createdAtInfo(timeStamp int64) (created_at string, created_at_week string, created_at_month string) {

	fmt.Println("Original time stamp is: ", timeStamp)
	t := time.Unix(0, timeStamp*int64(time.Millisecond))
	fmt.Println("Converted time is: ", t)

	created_at = t.Format(ISOformat)
	fmt.Println("The formatted YYYYMMDD created_at is", created_at)

	year, week := t.ISOWeek()
	created_at_week = strconv.Itoa(year) + "W" + strconv.Itoa(week)
	fmt.Println("The formatted YYYYWww created_at_week is ", created_at_week)

	created_at_month = t.Format(ISOMonthFormat)
	fmt.Println("The formatted YYYYMM created_month is ", created_at_month)

	return created_at, created_at_week, created_at_month
}
