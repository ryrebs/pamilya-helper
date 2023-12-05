package utils

import "strings"

func IsErrorMSG(msg string) bool { return strings.Contains(msg, "wrong") }
