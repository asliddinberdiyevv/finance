package utils

import (
	"net/url"
	"time"
)

// TimeParam - get time value from request query
func TimeParam(query url.Values, name string) (time.Time, error) {
	// return NOW as default time
	t := time.Now()
	value := query.Get(name)

	if value == "" {
		return t, nil
	}

	parsed, err := time.Parse(time.RFC3339, value)
	if err != nil {
		return t, err
	}

	return parsed, nil
}
