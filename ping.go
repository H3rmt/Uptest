package main

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"regexp"
	"strings"
	"time"
)

func CheckUrl(uri string, regex string) {
	log.Printf("Checking %s", uri)

	currentTime := time.Now().Format("2006-01-02 15:04:05")
	escapedURI := regexp.MustCompile(`[^a-zA-Z0-9-_.]`).ReplaceAllString(uri, "_") // remove all non-alphanumeric characters

	data, responseTime, errPing := ping(uri, regex)
	if errPing != nil {
		log.Printf("Error: %v", errPing)
	}

	if _, err := os.Stat(fmt.Sprintf("logs/%s.log", escapedURI)); os.IsNotExist(err) {
		logFile, err := os.Create(fmt.Sprintf("logs/%s.log", escapedURI))
		if err != nil {
			log.Printf("Error creating log file: %v", err)
		}
		defer logFile.Close()

		logFile.WriteString(fmt.Sprintf("%s|%s\n", uri, regex))
	}
	logFile, err := os.OpenFile(fmt.Sprintf("logs/%s.log", escapedURI), os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		log.Printf("Error opening log file: %v", err)
	}
	defer logFile.Close()

	logFile.WriteString(fmt.Sprintf("%s| %-7s", currentTime, fmt.Sprintf("%5dms", responseTime.Milliseconds())))
	if errPing != nil {
		_, err := os.Stat(fmt.Sprintf("responses/%s.d", escapedURI))
		if os.IsNotExist(err) {
			err := os.Mkdir(fmt.Sprintf("responses/%s.d", escapedURI), 0755)
			if err != nil {
				log.Printf("Error creating responses directory: %v", err)
			}
		}

		path := fmt.Sprintf("responses/%s.d/%s.html", escapedURI, currentTime)
		err = os.WriteFile(path, data, 0644)
		if err != nil {
			log.Printf("Error writing response to file: %v", err)
		}
		logFile.WriteString(fmt.Sprintf(" > [%s]", errPing))
	} else {
		path := fmt.Sprintf("responses/%s.html", escapedURI)
		err = os.WriteFile(path, data, 0644)
		if err != nil {
			log.Printf("Error writing response to file: %v", err)
		}
		logFile.WriteString(" > [OK]")
	}
	logFile.WriteString("\n")
}

func ping(url string, regex string) ([]byte, time.Duration, error) {
	// If the url does not start with "http:" or "https:", default to "https:".
	if !strings.HasPrefix(url, "http:") && !strings.HasPrefix(url, "https:") {
		url = "https://" + url
	}

	start := time.Now()

	res, err := http.Get(url)
	if err != nil {
		elapsed := time.Since(start)
		return nil, elapsed, err
	}

	elapsed := time.Since(start)

	body, err := io.ReadAll(res.Body)
	if err != nil {
		return nil, elapsed, err
	}

	if !strings.Contains(string(body), regex) {
		return body, elapsed, fmt.Errorf("regex %s not found in %s", regex, url)
	}

	return body, elapsed, nil
}
