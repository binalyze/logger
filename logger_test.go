package logger

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"io/ioutil"
	"math/rand"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/require"
)

type testData struct {
	Level   logrus.Level
	Time    time.Time
	Version string
	Message string
	File    string
}

var data = &testData{
	Level:   logrus.DebugLevel,
	Time:    time.Now(),
	Version: appVersion,
	Message: "Test Message",
	File:    "file:main.go:33",
}

// TestInit Success
func TestFormatter(t *testing.T) {

	mockEntry := logrus.Entry{
		Message: data.Message,
		Time:    data.Time,
		Level:   data.Level,
		Data:    logrus.Fields{"file": "main.go:33"},
	}

	f := formatter{}
	actual, err := f.Format(&mockEntry)
	require.NoError(t, err)

	// Example expected: "DEBUG 2021-01-23T14:43:03+03:00 1.0.0 Test Message main.go:33\n"
	expected := fmt.Sprintf("%s %s %s %s %s",
		convertLevel(data.Level),
		data.Time.Format(time.RFC3339),
		appVersion,
		data.Message,
		data.File,
	)

	require.Contains(t, string(actual), expected)

}

func TestSetOutputFile(t *testing.T) {

	f, err := ioutil.TempFile("", "_logger_set_output_*")
	require.NoError(t, err)
	defer func() {
		os.Remove(f.Name())
	}()
	logFile = f.Name()

	err = Init()
	require.NoError(t, err)
	logger.Out = getWriter()

	message := randStringBytes(30)

	Errorf("%s", message)

	content, err := ioutil.ReadFile(f.Name())
	require.NoError(t, err)

	split := strings.Split(string(content), " ")

	require.Equal(t, message, split[3])
}

func TestSetOutputConsole(t *testing.T) {

	// Create temp log file
	f, err := ioutil.TempFile("", "_logger_set_output_*")
	require.NoError(t, err)
	defer func() {
		os.Remove(f.Name())
	}()

	// Mock data
	logFile = f.Name()
	os.Setenv(envLogToConsole, "true")
	message := randStringBytes(30)

	// Redirect stdout to pipe
	old := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	// Init with os.Stdout and file as writer
	err = Init()
	require.NoError(t, err)
	logger.Out = getWriter()

	// Log random generated message
	Errorf("%s", message)

	// Catch stdout content from pipe
	outC := make(chan string)
	go func() {
		var buf bytes.Buffer
		_, err := io.Copy(&buf, r)
		require.NoError(t, err)
		outC <- buf.String()

	}()

	// Close pipe
	_ = w.Close()

	// Get content from channel
	content := <-outC

	// Reset stdout
	os.Stdout = old

	// Compare stdout content and message
	split := strings.Split(string(content), " ")
	require.Equal(t, message, split[3])
}

func TestLogFatal(t *testing.T) {

	// Mock data
	f, err := ioutil.TempFile("", "_logger_set_output_*")
	require.NoError(t, err)
	defer func() {
		os.Remove(f.Name())
	}()
	message := randStringBytes(30)
	logFile = f.Name()

	err = Init()
	require.NoError(t, err)

	logger.Out = getWriter()

	old := logger.ExitFunc
	defer func() {
		logger.ExitFunc = old
	}()

	var exitCode int
	exitter := func(code int) {
		exitCode = code
	}

	logger.ExitFunc = exitter

	Fatalf(message)

	require.Equal(t, 1, exitCode)

	content, err := ioutil.ReadFile(f.Name())
	require.NoError(t, err)

	require.Contains(t, string(content), message)

}

func TestLoggerHelpersDebugDisabled(t *testing.T) {

	f, err := ioutil.TempFile("", "_logger_set_output_*")
	require.NoError(t, err)
	defer func() {
		os.Remove(f.Name())
	}()

	// Mock data
	logFile = f.Name()
	os.Unsetenv(envLogToConsole)

	err = Init()
	require.NoError(t, err)

	logger.Out = getWriter()
	SetDebugLogging(false)

	messageDebug := randStringBytes(30)
	messageInfo := randStringBytes(30)
	messageWarning := randStringBytes(30)
	messageError := randStringBytes(30)

	Debugf("%s", messageDebug)
	Infof("%s", messageInfo)
	Warnf("%s", messageWarning)
	Errorf("%s", messageError)

	scanner := bufio.NewScanner(f)
	count := 0
	for scanner.Scan() {
		split := strings.Split(scanner.Text(), " ")
		switch count {
		// At line 0, there is set debug logging info message.
		case 1:
			require.Equal(t, "INFO", split[0])
			require.Equal(t, messageInfo, split[3])
		case 2:
			require.Equal(t, "WARNING", split[0])
			require.Equal(t, messageWarning, split[3])
		case 3:
			require.Equal(t, "ERROR", split[0])
			require.Equal(t, messageError, split[3])
		}

		count++
	}

	if err := scanner.Err(); err != nil {
		t.Errorf("Bufio scanner error: %v", err)
	}

}

func TestLoggerHelpersDebugEnabled(t *testing.T) {

	f, err := ioutil.TempFile("", "_logger_set_output_*")
	require.NoError(t, err)
	defer func() {
		os.Remove(f.Name())
	}()

	// Mock data
	logFile = f.Name()
	os.Unsetenv(envLogToConsole)

	err = Init()
	require.NoError(t, err)
	logger.Out = getWriter()
	SetDebugLogging(true)

	messageDebug := randStringBytes(30)
	messageInfo := randStringBytes(30)
	messageWarning := randStringBytes(30)
	messageError := randStringBytes(30)

	Debugf("%s", messageDebug)
	Infof("%s", messageInfo)
	Warnf("%s", messageWarning)
	Errorf("%s", messageError)

	scanner := bufio.NewScanner(f)
	count := 0
	for scanner.Scan() {
		split := strings.Split(scanner.Text(), " ")
		switch count {
		case 1:
			require.Equal(t, "DEBUG", split[0])
			require.Equal(t, messageDebug, split[3])
		case 2:
			require.Equal(t, "INFO", split[0])
			require.Equal(t, messageInfo, split[3])
		case 3:
			require.Equal(t, "WARNING", split[0])
			require.Equal(t, messageWarning, split[3])
		case 4:
			require.Equal(t, "ERROR", split[0])
			require.Equal(t, messageError, split[3])
		}

		count++
	}

	if err := scanner.Err(); err != nil {
		t.Errorf("Bufio scanner error: %v", err)
	}

}

func TestWrite(t *testing.T) {
	w := Writer()
	require.NotNil(t, w)
}

func convertLevel(level logrus.Level) string {
	levelMap := map[logrus.Level]string{
		logrus.PanicLevel: "PANIC",
		logrus.FatalLevel: "FATAL",
		logrus.ErrorLevel: "ERROR",
		logrus.WarnLevel:  "WARN",
		logrus.InfoLevel:  "INFO",
		logrus.DebugLevel: "DEBUG",
		logrus.TraceLevel: "TRACE",
	}

	return levelMap[level]
}

func randStringBytes(n int) string {
	const letterBytes = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"

	b := make([]byte, n)
	for i := range b {
		b[i] = letterBytes[rand.Intn(len(letterBytes))]
	}
	return string(b)
}
