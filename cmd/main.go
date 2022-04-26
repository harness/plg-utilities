package main

import (
	"fmt"
	stackdriver "github.com/TV4/logrus-stackdriver-formatter"
	"github.com/sirupsen/logrus"
	"github.com/spf13/viper"
	"os"
	. "plg-utilities/config"
	"plg-utilities/jobs"
)

func main() {
	// setup logger
	initLog()

	// set up app configuration
	config, err := initConfig()
	if err != nil {
		logrus.Fatalf("could not load configuartion, error: %s", err.Error())
	}
	fmt.Printf("%v\n", config)

	// run jobs
	jobs.RunJobs(&config)
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
