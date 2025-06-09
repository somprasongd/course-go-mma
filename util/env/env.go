package env

import (
	"os"
	"strconv"
	"time"
)

func Get(key string) string {
	v, ok := os.LookupEnv(key)
	if !ok {
		return ""
	}
	return v
}

func GetDefault(key string, defaultValue string) string {
	v, ok := os.LookupEnv(key)
	if !ok {
		return defaultValue
	}
	return v
}

func GetInt(key string) int {
	v, err := strconv.Atoi(Get(key))
	if err != nil {
		return 0
	}
	return v
}

func GetIntDefault(key string, defaultValue int) int {
	v, err := strconv.Atoi(Get(key))
	if err != nil {
		return defaultValue
	}
	return v
}
func GetFloat(key string) float64 {
	v, err := strconv.ParseFloat(Get(key), 64)
	if err != nil {
		return 0.0
	}
	return v
}

func GetFloatDefault(key string, defaultValue float64) float64 {
	v, err := strconv.ParseFloat(Get(key), 64)
	if err != nil {
		return defaultValue
	}
	return v
}

func GetBool(key string) bool {
	v := Get(key)
	switch v {
	case "true", "yes":
		return true
	case "false", "no":
		return false
	default:
		return false
	}
}
func GetBoolDefault(key string, defaultValue bool) bool {
	v := Get(key)
	switch v {
	case "true", "yes":
		return true
	case "false", "no":
		return false
	default:
		return defaultValue
	}
}

func GetDuration(key string) time.Duration {
	v := Get(key)
	if len(v) == 0 {
		return 0
	}
	d, err := time.ParseDuration(v)
	if err != nil {
		return 0
	}
	return d
}

func GetDurationDefault(key string, defaultValue time.Duration) time.Duration {
	v := Get(key)
	if len(v) == 0 {
		return defaultValue
	}
	d, err := time.ParseDuration(v)
	if err != nil {
		return defaultValue
	}
	return d
}
