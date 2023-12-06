package utils

import (
	"log"
	"strings"
)

// Check if msg is an error Message
func IsErrorMSG(msg string) bool { return strings.Contains(msg, "wrong") }

// Create a | separated string from an array of string.
func CreateListString(str []string) string {
	filtered := []string{}
	resp := ""

	// Remove empty strings
	for _, v := range str {
		if v != "" {
			filtered = append(filtered, v)
		}
	}

	// Append |
	for idx, v := range filtered {
		log.Println(idx, len(str)-1)
		if idx == len(filtered)-1 {
			resp = resp + v
		} else {
			resp = resp + v + "|"
		}
	}

	return resp
}
