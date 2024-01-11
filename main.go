package main

import (
	"bufio"
	"fmt"
	"html/template"
	"io"
	"log"
	"net/http"
	"os"
	"regexp"
	"strings"
	"sync"
	"time"
)

func main() {
	_, err := os.Stat("responses")
	if os.IsNotExist(err) {
		err := os.Mkdir("responses", 0755)
		if err != nil {
			log.Printf("Error creating responses directory: %v", err)
		}
	}

	_, err = os.Stat("logs")
	if os.IsNotExist(err) {
		err := os.Mkdir("logs", 0755)
		if err != nil {
			log.Printf("Error creating logs directory: %v", err)
		}
	}

	urls := os.Getenv("URLS")
	if urls == "" {
		log.Fatal("No URLS environment variable found")
	}

	http.HandleFunc("/favicon.ico", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, "favicon.ico")
	})

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		logs := make(map[Site][]Log)
		files, err := os.ReadDir("logs")
		if err != nil {
			log.Printf("Error reading logs directory: %v", err)
		}
		for _, file := range files {
			logFile, err := os.Open(fmt.Sprintf("logs/%s", file.Name()))
			if err != nil {
				log.Printf("Error opening log file: %v", err)
			}
			defer logFile.Close()

			names := strings.Split(file.Name(), ".")
			names = names[:len(names)-1]
			name := strings.Join(names, ".")

			fileScanner := bufio.NewScanner(logFile)
			fileScanner.Split(bufio.ScanLines)

			fileScanner.Scan()
			line := fileScanner.Text()
			url := strings.TrimSpace(strings.Split(line, "|")[0])
			regex := strings.TrimSpace(strings.Split(line, "|")[1])

			site := Site{
				Name:  name,
				Url:   url,
				Regex: regex,
			}

			logs[site] = make([]Log, 0)

			for fileScanner.Scan() {
				line := fileScanner.Text()

				time := strings.TrimSpace(strings.Split(line, "|")[0])
				delay := strings.TrimSpace(strings.Split(strings.Split(line, "|")[1], ">")[0])
				err := strings.TrimSpace(strings.Split(strings.Split(line, "|")[1], ">")[1])

				logs[site] = append(logs[site], Log{
					Time:  time,
					Delay: delay,
					Error: err,
				})
			}
		}

		tmpl := template.Must(template.ParseFiles("index.html"))
		tmpl.Execute(w, logs)
	})

	http.HandleFunc("/check", func(w http.ResponseWriter, r *http.Request) {
		var wg sync.WaitGroup
		for _, url := range strings.Split(urls, ",") {
			urlRegex := strings.Split(url, "=")
			if len(urlRegex) != 2 {
				log.Fatalf("Invalid URLS environment variable: %s", url)
			}
			wg.Add(1)
			go func() {
				defer wg.Done()
				CheckUrl(urlRegex[0], urlRegex[1])
			}()
		}
		wg.Wait()

		http.Redirect(w, r, "/", http.StatusSeeOther)
	})

	log.Fatal(http.ListenAndServe(":80", nil))
}

type Site struct {
	Name  string
	Url   string
	Regex string
}

// data struct for one log
type Log struct {
	Time  string
	Delay string
	Error string
}

func CheckUrl(uri string, regex string) {
	log.Printf("Checking %s", uri)

	currentTime := time.Now().Format("2006-01-02 15:04:05")
	escapedURI := regexp.MustCompile(`[^a-zA-Z0-9-_.]`).ReplaceAllString(uri, "_") // remove all non-alphanumeric characters

	data, responseTime, errPing := Ping(uri, regex)
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
		path := fmt.Sprintf("responses/%s.html", currentTime)
		err = os.WriteFile(path, data, 0644)
		if err != nil {
			log.Printf("Error writing response to file: %v", err)
		}
		logFile.WriteString(fmt.Sprintf(" > [%s] (%s)", errPing, path))
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

func Ping(url string, regex string) ([]byte, time.Duration, error) {
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
