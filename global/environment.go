package global

import "os"

func OtelEnabled() bool {
	return os.Getenv("STANZA_NO_OTEL") == ""
}

func SentinelEnabled() bool {
	return os.Getenv("STANZA_NO_SENTINEL") == ""
}
