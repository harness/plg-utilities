package segment

import "github.com/sirupsen/logrus"

type Logger interface {
	// Logf Segment clients call this method to log regular messages about the
	// operations they perform.
	// Messages logged by this method are tagged with an `INFO` log level
	Logf(format string, args ...interface{})

	// Errorf Analytics clients call this method to log errors they encounter while
	// sending events to the backend servers.
	// Messages logged by this method are usually tagged with an `ERROR` log level
	Errorf(format string, args ...interface{})
}

// DefaultSegmentLogger Default logger implementation using logrus lib
type DefaultSegmentLogger struct {
}

func (*DefaultSegmentLogger) Logf(format string, args ...interface{}) {
	logrus.Infof(format, args...)
}

func (*DefaultSegmentLogger) Errorf(format string, args ...interface{}) {
	logrus.Errorf(format, args...)
}
