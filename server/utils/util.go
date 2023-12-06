package utils

import (
	"io"
	"log"
	"mime/multipart"
	"os"
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

// Remove file if exists
func RemoveFile(fname string) error {
	if _, err := os.Stat(fname); err != nil {
		log.Println(err)
		return err
	}
	err := os.Remove(fname)
	if err != nil {
		log.Println(err)
		return err
	}
	return nil
}

// Create file uploads
func CreateFile(file *multipart.FileHeader, newFilename string) error {
	src, err := file.Open()
	if err != nil {
		return err
	}
	defer src.Close()
	dst, err := os.Create(newFilename)
	if err != nil {
		log.Println(err)
		return err
	}
	defer dst.Close()
	if _, err = io.Copy(dst, src); err != nil {
		return err
	}
	return nil
}
