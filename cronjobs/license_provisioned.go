package cronjobs

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

const (
	CD_LICENSE_UNIT                          = "cd_license_unit"
	CD_LICENSE_WORKLOADS_PROVISIONED         = "cd_license_workloads_provisioned"
	CD_LICENSE_SERVICES_PROVISIONED          = "cd_license_services_provisioned"
	CD_LICENSE_SERVICE_INSTANCES_PROVISIONED = "cd_license_service_instances_provisioned"
	CI_LICENSE_DEVELOPERS_PROVISIONED        = "ci_license_developers_provisioned"
	FF_LICENSE_DEVELOPERS_PROVISIONED        = "ff_license_developers_provisioned"
	FF_LICENSE_MAU_PROVISIONED               = "ff_license_mau_provisioned"
	CCM_LICENSE_CLOUD_SPEND_PROVISIONED      = "ccm_license_cloud_spend_provisioned"

	CD                 = "CD"
	SERVICES           = "SERVICES"
	SERVICES_INSTANCES = "SERVICE_INSTANCES"

	CI = "CI"
	CF = "CF"
	CE = "CE"

	SERVICES_KEY           = "Services"
	SERVICES_INSTANCES_KEY = "Service Instances"
	CUSTOM                 = "Custom"
	UNDEFINED              = "Undefined"
)

func RunLicenseProvisionedJob(mongo *mongodb.MongoDb, segmentSender *segment.HTTPClient) error {
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
	collectionName := mongo.ModuleLicenseDAO.ModuleLicenseCollection.Name()
	moduleLicenseCursor, err := mongo.ModuleLicenseDAO.ListWithCursor(ctx)
	if err != nil {
		logrus.Fatalf("unable to list collection %s: %s", collectionName, err.Error())
	}
	defer moduleLicenseCursor.Close(ctx)
	for moduleLicenseCursor.Next(ctx) {
		var moduleLicense core.ModuleLicense
		err := moduleLicenseCursor.Decode(&moduleLicense)
		if err != nil {
			logrus.Errorf("unable to decode record for collection %s: %+v: %s", collectionName, moduleLicense, err.Error())
			continue
		}
		logrus.Infof("found in collection %s: %+v", collectionName, moduleLicense)
		createLicenseGroupEvent(moduleLicense, &batchEvents, batchEventsQueue)
	}

	// flush batch events if there is any left
	if len(batchEvents) != 0 {
		flush(&batchEvents, batchEventsQueue)
	}

	if err := moduleLicenseCursor.Err(); err != nil {
		logrus.Errorf("unable to list entire collection %s: %s", collectionName, err.Error())
	}

	close(batchEventsQueue)
	wg.Wait()
	return nil
}

func createLicenseGroupEvent(moduleLicense core.ModuleLicense, batchEvents *[]analytics.Message, queue chan []analytics.Message) {
	accountId := moduleLicense.AccountIdentifier
	traits := getTraits(moduleLicense, accountId)
	event := analytics.Group{
		UserId:    segment.ACCOUNT_ANALYSIS_USER_PREFIX + accountId,
		GroupId:   accountId,
		Timestamp: time.Now(),
		//todo: group_name might be a required trait for segment
		Traits:       traits,
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

func getTraits(moduleLicense core.ModuleLicense, accountId string) map[string]interface{} {
	traits := map[string]interface{}{"group_id": accountId, "group_type": "Account"}
	if moduleLicense.ModuleType == CD {
		if moduleLicense.Status == "ACTIVE" {
			traits[CD_LICENSE_WORKLOADS_PROVISIONED] = moduleLicense.Workloads
			switch moduleLicense.CDLicenseType {
			case SERVICES:
				traits[CD_LICENSE_UNIT] = SERVICES_KEY
			case SERVICES_INSTANCES:
				traits[CD_LICENSE_UNIT] = SERVICES_INSTANCES_KEY
			default:
				traits[CD_LICENSE_UNIT] = CUSTOM
			}
		} else {
			traits[CD_LICENSE_WORKLOADS_PROVISIONED] = UNDEFINED
			traits[CD_LICENSE_UNIT] = UNDEFINED
		}
	}

	if moduleLicense.ModuleType == CI {
		traits[CI_LICENSE_DEVELOPERS_PROVISIONED] = moduleLicense.NumberOfCommitters
	}

	if moduleLicense.ModuleType == CF {
		traits[FF_LICENSE_DEVELOPERS_PROVISIONED] = moduleLicense.NumberOfUsers
		traits[FF_LICENSE_MAU_PROVISIONED] = moduleLicense.NumberOfClientMAUs
	}

	if moduleLicense.ModuleType == CE {
		traits[CCM_LICENSE_CLOUD_SPEND_PROVISIONED] = moduleLicense.SpendLimit
	}

	return traits
}
