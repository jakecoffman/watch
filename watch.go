package main

import (
	"fmt"
	"log"
	"os/exec"
	"time"

	"code.google.com/p/go.exp/fsnotify"
)

func main() {
	log.Println("Press Ctl-C to stop watching\n")

	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		log.Fatal(err)
	}

	// run once when first starting up
	run()

	go func() {
		for {
			select {
			case <-watcher.Event:
				// watcher's events sometimes fire rapidly, so introduce a 100ms delay before running
				time.Sleep(time.Millisecond * 100)
				drainChan(watcher.Event)
				run()
			case err := <-watcher.Error:
				log.Fatal("error:", err)
			}
		}
	}()

	// watches cwd
	err = watcher.Watch(".")
	if err != nil {
		log.Fatal(err)
	}

	// block forever (or until a log.Fatal)
	<-make(chan bool)
}

// do a run of the command (hard coded to `go test` for now)
func run() {
	log.Println("Starting run")

	cmd := exec.Command("go", "test")
	out, err := cmd.CombinedOutput()
	fmt.Print(string(out))

	if err != nil {
		log.Printf("%v\n\n", err)
	} else {
		log.Println("Run complete\n")
	}
}

// takes a channel and reads everything from it until it is empty and then returns
func drainChan(c chan *fsnotify.FileEvent) {
	for {
		select {
		case <-c:
		default:
			return
		}
	}
}
