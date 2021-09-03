package logger

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
	"time"

	"github.com/sirupsen/logrus"
	"gopkg.in/natefinch/lumberjack.v2"
)

const (
	skipFrameCount    = 4
	splitAfterPkgName = "github.com/binalyze/logger"

	envLogToConsole = "LOG_TO_CONSOLE"

	maxSizeInMBs         = 10
	maxBackups           = 3
	maxAgeInDays         = 30
	enableLogCompression = true
)

var (
	appVersion = "1.0.0"

	logger  = logrus.New()
	logFile = getLogFileName(".log")
)

// Init initiates logger with writer, formatter and level
func Init() error {
	logger.Out = getWriter()
	logger.Formatter = &formatter{}
	logger.Level = logrus.InfoLevel

	return nil
}

// SetPrefix prepends prefix s to the log messages and call it thread safe.
func SetPrefix(s string) {
	logger.SetFormatter(&formatter{prefix: s})
}

// Debugf logs a message at level Debug on the standard logger.
func Debugf(format string, args ...interface{}) {
	if logger.IsLevelEnabled(logrus.DebugLevel) {
		entry := newEntry()
		entry.Debugf(format, args...)
	}
}

// Infof logs a message at level Info on the standard logger.
func Infof(format string, args ...interface{}) {
	entry := newEntry()
	entry.Infof(format, args...)
}

// Warnf logs a message at level Warn on the standard logger.
func Warnf(format string, args ...interface{}) {
	entry := newEntry()
	entry.Warnf(format, args...)
}

// Errorf logs a message at level Error on the standard logger.
func Errorf(format string, args ...interface{}) {
	entry := newEntry()
	entry.Errorf(format, args...)
}

// Fatalf logs a message at level Fatal on the standard logger.
func Fatalf(format string, args ...interface{}) {
	entry := newEntry()
	entry.Fatalf(format, args...)
}

// Writer returns the underlying io.Writer instance of the logger.
func Writer() io.Writer {
	return logger.Out
}

// SetDebugLogging sets the logging level
func SetDebugLogging(enabled bool) {
	logger.Infof("Debug logging set to: %t", enabled)

	if enabled {
		logger.SetLevel(logrus.DebugLevel)
		return
	}

	// If not enabled, set to default info level
	logger.SetLevel(logrus.InfoLevel)
}

// GetLevel returns the logger instance's log level and exported for testing purposes to determine log level is set
// correctly.
func GetLevel() logrus.Level {
	return logger.GetLevel()
}

// newEntry creates new logrus Entry with logrus fields, file, line and function
func newEntry() *logrus.Entry {
	file, function, line := callerInfo(skipFrameCount, splitAfterPkgName)

	entry := logger.WithFields(logrus.Fields{})
	entry.Data["file"] = file
	entry.Data["line"] = line
	entry.Data["function"] = function
	return entry
}

// callerInfo grabs caller file, function and line number
func callerInfo(skip int, pkgName string) (file, function string, line int) {

	// Grab frame
	pc := make([]uintptr, 1)
	n := runtime.Callers(skip, pc)
	frames := runtime.CallersFrames(pc[:n])
	frame, _ := frames.Next()

	// Set file, function and line number
	file = trimPkgName(frame.File, pkgName)
	function = trimPkgName(frame.Function, pkgName)
	line = frame.Line

	return
}

// trimPkgName trims string after splitStr
func trimPkgName(frameStr, splitStr string) string {
	count := strings.LastIndex(frameStr, splitStr)
	if count > -1 {
		frameStr = frameStr[count+len(splitStr):]
	}

	return frameStr
}

func getWriter() io.Writer {
	logToConsole := os.Getenv(envLogToConsole) != ""

	// Set output according to environment variable
	var output io.Writer
	if logToConsole {
		output = io.MultiWriter(getRotatedFile(), os.Stdout)
	} else {
		output = getRotatedFile()
	}

	return output
}

func getLogFileName(extension string) string {
	appName := filepath.Base(os.Args[0])
	ext := filepath.Ext(appName)
	fmt.Println(ext)
	if len(ext) > 0 {
		return strings.Replace(appName, ext, extension, 1)
	}

	// No extension in filename
	return appName + extension
}

// getRotatedFile sets the output to desired file
func getRotatedFile() io.Writer {
	return &lumberjack.Logger{
		Filename:   logFile,
		MaxSize:    maxSizeInMBs,
		MaxBackups: maxBackups,
		MaxAge:     maxAgeInDays,
		Compress:   enableLogCompression,
	}
}

// Formatter implements logrus.Formatter interface.
type formatter struct {
	prefix string
}

// Format building log message.
func (f *formatter) Format(entry *logrus.Entry) ([]byte, error) {
	var sb bytes.Buffer

	var newLine = "\n"
	if runtime.GOOS == "windows" {
		newLine = "\r\n"
	}

	sb.WriteString(strings.ToUpper(entry.Level.String()))
	sb.WriteString(" ")
	sb.WriteString(entry.Time.Format(time.RFC3339))
	sb.WriteString(" ")
	sb.WriteString(appVersion)
	sb.WriteString(" ")
	sb.WriteString(f.prefix)
	sb.WriteString(entry.Message)
	sb.WriteString(" ")
	file, ok := entry.Data["file"].(string)
	if ok {
		sb.WriteString("file:")
		sb.WriteString(file)
	}
	line, ok := entry.Data["line"].(int)
	if ok {
		sb.WriteString(":")
		sb.WriteString(strconv.Itoa(line))
	}
	function, ok := entry.Data["function"].(string)
	if ok {
		sb.WriteString(" ")
		sb.WriteString("func:")
		sb.WriteString(function)
	}
	sb.WriteString(newLine)

	return sb.Bytes(), nil
}
