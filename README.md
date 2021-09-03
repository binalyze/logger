# logger
logger package is a log management tool which has helper functions to log into file.

## Usage
There is only one function to initialize logger.

`logger.Init()`

When the logger package is initialized with logger.Init, user can log with the helper functions below.

```go
// Get flag option values
logger.Debugf("%s","This is a debug log")
logger.Infof("%s","This is a debug log")
logger.Warnf("%s","This is a debug log")
logger.Errorf("%s","This is a debug log")
logger.Fatalf("%s","This is a debug log")
```

**Example log:**
ERROR 2021-01-26T14:37:17+03:00 1.0.0 Test logging main.go:25

### Level
Logger package logs only Error and Fatal levels by default. If debug logging is enabled, then all the levels will be logged. The debug logging is automatically sets according to `config.GetDebugLogging()`. However it can be change anytime with the helper below:

`logger.EnableDebugLogging(true)`