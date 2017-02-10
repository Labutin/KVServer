package logs

type LogsLevels uint8

// Logs levels.
const (
	DEBUG = LogsLevels(iota)
	INFO
	WARN
	ERROR
)

// Logs Levels in string format
var logsLevels = map[LogsLevels]string{
	DEBUG: "DEBUG",
	INFO:  "INFO",
	WARN:  "WARN",
	ERROR: "ERROR",
}

func (v LogsLevels) String() string { return logsLevels[v] }

// MakeLogString format logs string
func MakeLogString(ll LogsLevels, goroutineID, errorMessage string, err error) string {
	errorString := "[" + ll.String() + "] "
	if goroutineID != "" {
		errorString += "(" + goroutineID + ") "
	}
	errorString += errorMessage
	if err != nil {
		errorString += " Error: " + err.Error()
	}

	return errorString
}
