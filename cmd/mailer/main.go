package main

import (
	"fmt"
	"io/fs"
	"log"
	"os"
	"time"

	"github.com/fsnotify/fsnotify"
	"github.com/josnelihurt/mailer-go/pkg/config"
	"github.com/josnelihurt/mailer-go/pkg/mailer"
)

func eventInfo(filename string) (fs.FileInfo, error) {
	file, err := os.Open(filename)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	fileInfo, err := file.Stat()
	if err != nil {
		return nil, err
	}

	return fileInfo, nil
}

func processFSEvent(config config.Config, events <-chan fsnotify.Event, errors <-chan error) {
	log.Printf("waitting for events")
	for {
		select {
		case event, ok := <-events:
			if !ok {
				return
			}
			time.Sleep(10 * time.Millisecond) //IDK it the other process writes immediatly
			if event.Op != fsnotify.Create {
				continue
			}
			fileToSend, err := eventInfo(event.Name)
			if err != nil {
				log.Println("unkown err: ", err)
				continue
			}
			mailer.SendFile(config, fileToSend)
		case err, ok := <-errors:
			if !ok {
				return
			}
			log.Println("error:", err)
		}
	}
}

func main() {
	fmt.Println("Starting mailer-go in client mode")
	config, err := config.Read()
	if err != nil {
		log.Fatal("Failed to read config:", err)
	}

	log.Printf("\nconfig ok: %v", config.String())

	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		log.Fatal("NewWatcher failed: ", err)
	}
	defer watcher.Close()

	done := make(chan bool)
	defer close(done)
	go processFSEvent(config, watcher.Events, watcher.Errors)

	if _, err := os.Stat(config.Inbox); os.IsNotExist(err) {
		log.Fatal("inbox folder doesn't exist", err)
	}
	if _, err := os.Stat(config.ErrBox); os.IsNotExist(err) {
		log.Fatal("err folder doesn't exist", err)
	}
	if _, err := os.Stat(config.DoneBox); os.IsNotExist(err) {
		log.Fatal("done folder doesn't exist", err)
	}

	files, err := os.ReadDir(config.Inbox)
	if err != nil {
		log.Fatal(err)
	}

	for _, file := range files {
		log.Printf("\n sending %s", file)
		fileInfo, err := file.Info()
		if err != nil {
			log.Fatal("unable to get info", err)
		}
		mailer.SendFile(config, fileInfo)
	}
	err = watcher.Add(config.Inbox)
	if err != nil {
		log.Fatal("Add failed:", err)
	}
	<-done
}
