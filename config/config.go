package config

import "os"

// GetEnv retrieves the value of an environment variable with a fallback value if not set
func GetEnv(key, fallback string) string {
	if value, exists := os.LookupEnv(key); exists {
		return value
	}
	return fallback
}
