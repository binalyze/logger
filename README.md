# logger
logger package is a log management tool which has helper functions to log into file.

## Usage
There is only one function to initialize logger.

`logger.Init()`

When the logger package is initialized with logger.Init, user can log with the helper functions below.

```go
// Get flag option values
logger.Debugf("%s","This is a debug log")
logger.Infof("%s","This is an info log")
logger.Warnf("%s","This is a warn log")
logger.Errorf("%s","This is an error log")
logger.Fatalf("%s","This is a fatal log")
```

**Example log:**
ERROR 2021-01-26T14:37:17+03:00 1.0.0 Test logging main.go:25

### Level
Logger package default log level is `Info`. If Debug logging is enabled, then all the levels will be logged. You can set log level to Debug with the helper below:

`logger.SetDebugLogging(true)`