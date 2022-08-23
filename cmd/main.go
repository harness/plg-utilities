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
	job "plg-utilities/jobs"
	"plg-utilities/telemetry/segment"
	"strings"
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

	jobsToDo := strings.Split(config.Mode, ",")
	jobs := make(map[string]bool)
	for _, job := range jobsToDo {
		fmt.Printf("%s\n", job)
		jobs[job] = true
	}

	mongo, err := mongodb.New(config.CGMongoDb, config.NGMongoDb)
	if err != nil {
		logrus.Fatalf("unable to connect to mongo db: %s", err.Error())
	}

	segmentSender := segment.NewHTTPClient(config.Segment)

	if jobs["LICENSE_PROVISIONED_CRON"] {
		cronjobs.RunLicenseProvisionedJob(mongo, segmentSender)
	}

	if jobs["ACCOUNT_TRAITS_CRON"] {
		fmt.Printf("ADASDASDAS\n")
		cronjobs.RunAccountTraitsJob(mongo, segmentSender)
	}

	if jobs["ANALYTICS_USER_JOB"] {
		job.RunJobs(&config)
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
