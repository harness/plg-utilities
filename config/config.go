package config

// Config provides the system configuration
type (
	Config struct {
		MongoDb MongoDbConf `mapstructure:",squash"`
		Segment SegmentConf `mapstructure:",squash"`
		Mode    string      `mapstructure:"MODE" default:"LICENSE_PROVISIONED_CRON"`
	}

	MongoDbConf struct {
		DbName  string `mapstructure:"CG_MONGODB_DB_NAME" default:"harness"`
		ConnStr string `mapstructure:"CG_MONGODB_URL" default:"mongodb://localhost:27017/harness"`
	}

	SegmentConf struct {
		Enabled bool   `mapstructure:"SEGMENT_ENABLED"`
		ApiKey  string `mapstructure:"SEGMENT_API_KEY"`
		Url     string `mapstructure:"SEGMENT_URL"`
	}
)
