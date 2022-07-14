package main

import (
	"fmt"
	stackdriver "github.com/TV4/logrus-stackdriver-formatter"
	"github.com/sirupsen/logrus"
	"github.com/spf13/viper"
	"os"
	. "plg-utilities/config"
	"plg-utilities/cronjobs"
	"plg-utilities/db/mongodb"
	"plg-utilities/jobs"
	"plg-utilities/telemetry/segment"
)

func main() {
	// setup logger
	initLog()

	// set up app configuration
	config, err := initConfig()
	if err != nil {
		logrus.Fatalf("could not load configuartion, error: %s", err.Error())
	}

	fmt.Printf("%+v\n", config)

	// run jobs
	if config.Mode == "ANALYTICS_USER_JOB" {
		jobs.RunJobs(&config)
	}

	if config.Mode == "LICENSE_PROVISIONED_CRON" {
		mongo, err := mongodb.New(config.CGMongoDb, config.NGMongoDb)
		if err != nil {
			logrus.Fatalf("unable to connect to mongo db: %s", err.Error())
		}
		segmentSender := segment.NewHTTPClient(config.Segment)
		cronjobs.RunLicenseProvisionedJob(mongo, segmentSender)
	}
}

func initLog() {
	level, ok := os.LookupEnv("LOG_LEVEL")
	if !ok {
		level = "info"
	}

	logLevel, err := logrus.ParseLevel(level)
	if err != nil {
		logLevel = logrus.InfoLevel
	}

	logrus.SetLevel(logLevel)
	logrus.SetFormatter(stackdriver.NewFormatter(
		stackdriver.WithService("plg-utilities"),
	))
}

// get config from env vars
func initConfig() (Config, error) {
	// add local path
	viper.AddConfigPath(".")
	// add another path for docker
	viper.AddConfigPath("/")

	// get config file
	viper.SetConfigName("config")
	viper.SetConfigType("yaml")
	viper.AutomaticEnv()

	cfg := Config{}
	err := viper.ReadInConfig()
	if err != nil {
		return cfg, err
	}
	err = viper.Unmarshal(&cfg)

	return cfg, err
}
