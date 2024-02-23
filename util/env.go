package util

import "os"

// Getenv returns the value of the environment variable named by
// the key, or fallback if the environment variable is not set.
func Getenv(key, fallback string) string {
	if value, ok := os.LookupEnv(key); ok {
		return value
	}
	return fallback
}
