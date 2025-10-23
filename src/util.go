package main

import "os"

func getEnvDefault(k, d string) string {
	if v := os.Getenv(k); v != "" { return v }
	return d
}
