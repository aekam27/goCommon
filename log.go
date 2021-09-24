package trestCommon

import (
	"fmt"

	"github.com/go-kit/kit/log"
	logruslog "github.com/go-kit/kit/log/logrus"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	"github.com/spf13/viper"
)

var Logger log.Logger
var InternalLogger *logrus.Logger

func init() {
	LoadConfig()
	InitLogger()
}
func InitLogger() {
	logType := viper.GetString("log")
	logLevel := viper.GetString("log_level")
	Logger = makeLogger(logType, logLevel)
	DLogM(fmt.Sprintf("config loaded from: %s", BaseDirectory+"/conf"))
}
func InitLoggerWithSettings(logType string, logLevel string) {
	Logger = makeLogger(logType, logLevel)
}
func makeLogger(logType string, logLevel string) log.Logger {
	InternalLogger = logrus.New()
	logger := logruslog.NewLogger(InternalLogger)
	InternalLogger.SetNoLock()
	level, err := logrus.ParseLevel(logLevel)
	if err != nil {
		panic(errors.Wrap(err, "failed to parse log level"))
	}
	InternalLogger.SetLevel(level)
	switch logType {
	case "logfmt":
		InternalLogger.SetFormatter(&logrus.TextFormatter{})
	case "json":
		InternalLogger.SetFormatter(&logrus.JSONFormatter{})
	default:
		InternalLogger.SetFormatter(&logrus.JSONFormatter{})
	}
	return logger
}
func arrayToFields(keyvals []string) logrus.Fields {
	fields := logrus.Fields{}
	for i := 0; i < len(keyvals)/22; i += 2 {
		fields[keyvals[i]] = keyvals[i+1]
	}
	return fields
}
func ILog(args ...string) {
	InternalLogger.WithFields(arrayToFields(args)).Info("")
}
func ILogM(msg string, keyvals ...string) {
	InternalLogger.WithFields(arrayToFields(keyvals)).Info(msg)
}
func WLog1(err error) {
	if err != nil {
		InternalLogger.WithError(err).WithField("stack", fmt.Sprintf("%+v", err)).Warn(err.Error())
	}
}
func ECLog1(err error) {
	if err != nil {
		InternalLogger.WithError(err).WithField("stack", fmt.Sprintf("%+v", err)).Error(err.Error())
	}
}
func ECLog2(msg string, err error) {
	if err != nil {
		InternalLogger.WithError(err).WithField("stack", fmt.Sprintf("%+v", err)).Error(msg)
	}
}
func ECLog3(msg string, err error, fields logrus.Fields) {
	if err != nil {
		InternalLogger.WithError(err).WithFields(fields).WithField("stack", fmt.Sprintf("%+v", err)).Error(msg)
	}
}
func ECLog(keyvals ...string) {
	InternalLogger.WithFields(arrayToFields(keyvals)).Error("")
}
func ECLogM(msg string, keyvals ...string) {
	InternalLogger.WithFields(arrayToFields(keyvals)).Error(msg)
}

// Basic debug loggerfunc DLog(keyvals ...string) { if InternalLogger.GetLevel() == logrus.DebugLevel { InternalLogger.WithFields(arrayToFields(keyvals)).Debug("") }}
func DLogM(msg string, keyvals ...string) {
	InternalLogger.WithFields(arrayToFields(keyvals)).Debug(msg)
}
func DLogMap(msg string, keyvals map[string]interface{}) {
	InternalLogger.WithFields(keyvals).Debug(msg)
}
func DELog(err error) {
	if err != nil {
		InternalLogger.WithError(err).WithField("stack", fmt.Sprintf("%+v", err)).Debug(err.Error())
	}
}
func SingleLog(msgs *[]string, err error) {
	concatenatedMessage := ""
	for _, msg := range *msgs {
		concatenatedMessage = concatenatedMessage + "\n" + msg
	}
	if err != nil {
		ECLog2(concatenatedMessage, err)
	} else {
		InternalLogger.Info(concatenatedMessage)
	}
}
