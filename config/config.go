package config

// Config provides the system configuration
type (
	Config struct {
		MongoDb MongoDbConf `mapstructure:",squash"`
		Segment SegmentConf `mapstructure:",squash"`
	}

	MongoDbConf struct {
		//DbName  string `mapstructure:"CG_MONGODB_DB_NAME" default:"harness"`
		//ConnStr string `mapstructure:"CG_MONGODB_URL" default:"mongodb://localhost:27017/harness"`
		DbName  string `mapstructure:"CG_MONGODB_DB_NAME"`
		ConnStr string `mapstructure:"CG_MONGODB_URL"`
		//EnableReflection bool   `envconfig:"GITOPS_SERVICE_MONGODB_ENABLE_REFLECTION" default:"true"`
	}

	SegmentConf struct {
		Enabled        bool   `mapstructure:"SEGMENT_ENABLED"`
		ApiKey         string `mapstructure:"SEGMENT_API_KEY"`
		CertRequired   bool   `mapstructure:"SEGMENT_CERT_REQUIRED"`
		BlockForEvents bool   `mapstructure:"SEGMENT_BLOCK_FOR_EVENTS"`
	}
)
