package jobs

import (
	"fmt"
	"github.com/sirupsen/logrus"
	"plg-utilities/config"
	"plg-utilities/db/mongodb"
	"plg-utilities/telemetry/segment"
)

//todo: convert this to migrations microservice
func RunJobs(config *config.Config) {
	fmt.Printf("%v\n", config)
	//logrus.Infof("%v\n", config.MongoDb)
	mongo, err := mongodb.New(config.MongoDb)
	if err != nil {
		logrus.Fatalf("unable to connect to mongo db: %s", err.Error())
	}

	segmentSender, _ := segment.New(config.Segment, SegmentLogger{})

	runAnalyticsUserCreate(mongo, segmentSender)
}

//Example logger implementation using logrus lib
type SegmentLogger struct {
}

func (SegmentLogger) Logf(format string, args ...interface{}) {
	logrus.Infof(format, args...)
}

func (SegmentLogger) Errorf(format string, args ...interface{}) {
	logrus.Errorf(format, args...)
}
