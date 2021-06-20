package main

import (
	"fmt"
	"os"
	"strings"
)

func envPanic(n string) string {
	result := os.Getenv(n)
	if result == "" {
		panic(fmt.Sprintf("Environment variable %s not defined", n))
	}
	return result
}

func escapeHtml(s string) string {
	result := strings.ReplaceAll(s, "<", "&gt;")
	result = strings.ReplaceAll(result, ">", "&lt;")
	result = strings.ReplaceAll(result, "&", "&amp;")
	return result
}
