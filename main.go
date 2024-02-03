package main

import (
	"log"
	"os"
)

func main() {
	check()
	RunHttp()
}

func check() {
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
}

type Site struct {
	Name       string
	Url        string
	EscapedUrl string
	Regex      string
}

// data struct for one log
type Log struct {
	Time  string
	Delay string
	Error string
}
