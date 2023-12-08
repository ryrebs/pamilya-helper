package utils

import (
	"io"
	"log"
	"mime/multipart"
	"os"
	"strings"

	"github.com/gorilla/sessions"
	"github.com/labstack/echo-contrib/session"
	"github.com/labstack/echo/v4"
)

// Check if msg is an error Message
func IsErrorMSG(msg string) bool { return strings.Contains(msg, "wrong") }

func FilterEmpty(str []string) []string {
	filtered := []string{}
	for _, v := range str {
		if v != "" {
			filtered = append(filtered, v)
		}
	}
	return filtered
}

// Create a | separated string from an array of string.
func CreateListString(str []string, filter func(str []string) []string) string {
	var filtered []string
	resp := ""

	if filter != nil {
		filtered = filter(str)
	} else {
		filtered = str
	}

	// Append |
	for idx, v := range filtered {
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

// Create session for flash message
func CreateFlashMessage(cc echo.Context, sessName string, maxAge int, path, msg, key string) *sessions.Session {
	sess, err := session.Get(sessName, cc)
	if err != nil {
		log.Println(err)
		return nil
	} else {
		sess.Options = &sessions.Options{
			MaxAge:   10,
			Path:     path,
			HttpOnly: true,
		}
		sess.AddFlash(msg, key)
		return sess
	}
}
