package pkg

import (
	"context"
)

func GetStateFromContext(c context.Context) string {
	return getValueFromContext(c, "state")
}

func GetCodeFromContext(c context.Context) string {
	return getValueFromContext(c, "code")
}

func getValueFromContext(c context.Context, key string) string {
	if c == nil {
		return ""
	}
	if v, ok := c.Value(key).(string); ok {
		return v
	}
	return ""
}
