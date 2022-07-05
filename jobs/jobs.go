package jobs

import (
	"github.com/sirupsen/logrus"
	"plg-utilities/config"
	"plg-utilities/db/mongodb"
	"plg-utilities/telemetry/segment"
)

// todo: convert this to migrations microservice
func RunJobs(config *config.Config) {
	mongo, err := mongodb.New(config.CGMongoDb, config.NGMongoDb)
	if err != nil {
		logrus.Fatalf("unable to connect to mongo db: %s", err.Error())
	}

	segmentSender := segment.NewHTTPClient(config.Segment)

	// job to create account analytics user
	runAnalyticsUserCreate(mongo, segmentSender)
}
