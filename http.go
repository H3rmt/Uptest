package main

import (
	"bufio"
	"fmt"
	"html/template"
	"io"
	"io/fs"
	"log"
	"net/http"
	"os"
	"regexp"
	"strings"
	"sync"
	"time"

	"github.com/labstack/echo"
)

func readLogs(files []fs.DirEntry) map[Site][]Log {
	logs := make(map[Site][]Log)
	for _, file := range files {
		logFile, err := os.Open(fmt.Sprintf("logs/%s", file.Name()))
		if err != nil {
			log.Printf("Error opening log file: %v", err)
			break
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
			Name:       name,
			Url:        url,
			EscapedUrl: regexp.MustCompile(`[^a-zA-Z0-9-_.]`).ReplaceAllString(url, "_"), // remove all non-alphanumeric characters
			Regex:      regex,
		}

		logs[site] = make([]Log, 0)

		for fileScanner.Scan() {
			line := fileScanner.Text()

			// remove from string until first pipe is found
			first := strings.SplitN(line, "|", 2)
			time := strings.TrimSpace(first[0])
			second := strings.SplitN(first[1], ">", 2)
			delay := strings.TrimSpace(second[0])
			err := strings.TrimSpace(second[1])

			logs[site] = append(logs[site], Log{
				Time:  time,
				Delay: delay,
				Error: err,
			})
		}
	}

	return logs
}

type Template struct {
	templates *template.Template
}

func (t *Template) Render(w io.Writer, name string, data interface{}, c echo.Context) error {
	return t.templates.ExecuteTemplate(w, name, data)
}

func RunHttp() {
	t := &Template{
		templates: template.Must(template.ParseFiles("index.html")),
	}

	e := echo.New()
	e.Renderer = t

	// serve files from responses directory
	e.Static("/responses", "responses")

	e.GET("/", func(c echo.Context) error {
		files, err := os.ReadDir("logs")
		if err != nil {
			log.Printf("Error reading logs directory: %v", err)
		}
		logs := readLogs(files)

		data := struct {
			Logs map[Site][]Log
			Time string
		}{
			Logs: logs,
			Time: time.Now().Format("2006-01-02 15:04:05"),
		}
		return c.Render(http.StatusOK, "index.html", data)
	})

	e.GET("/check", func(c echo.Context) error {
		var wg sync.WaitGroup
		for _, url := range strings.Split(os.Getenv("URLS"), ",") {
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

		// redirect to home page
		return c.Redirect(http.StatusSeeOther, "/")
	})

	e.Logger.Fatal(e.Start(":8080"))
}