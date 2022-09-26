package logging

import "os"

const DefaultLogLevel = "info"

func LogLevel() string {
	if str, ok := os.LookupEnv("LOGGING_LOGLEVEL"); ok {
		return str
	}

	return DefaultLogLevel
}

func isDev() bool {
	envString, ok := os.LookupEnv("LOGGING_ENVIRONMENT")

	if ok && envString == "dev" {
		return true
	}

	return false
}
