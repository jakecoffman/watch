package main

import (
	"fmt"
	"log"
	"os"
	"os/exec"
	"strings"
	"time"

	"code.google.com/p/go.exp/fsnotify"
)

func main() {
	cmd := parse()
	log.Printf("Running %v", strings.Join(cmd, " "))
	log.Println("Press Ctl-C to stop watching\n")

	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		log.Fatal(err)
	}

	// run once when first starting up
	run(cmd)

	go watcherHandler(watcher, cmd)

	// watches cwd
	err = watcher.Watch(".")
	if err != nil {
		log.Fatal(err)
	}

	// block forever (or until a log.Fatal)
	<-make(chan bool)
}

func parse() []string {
	var cmd []string
	if len(os.Args) == 1 {
		cmd = []string{"go", "test"}
	} else {
		cmd = os.Args[1:]
	}
	return cmd
}

// do a run of the command (hard coded to `go test` for now)
func run(cmd []string) {
	log.Println("Starting run")

	c := exec.Command(cmd[0], cmd[1:]...)
	out, err := c.CombinedOutput()
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

func watcherHandler(watcher *fsnotify.Watcher, cmd []string) {
	for {
		select {
		case <-watcher.Event:
			// watcher's events sometimes fire rapidly, so introduce a 100ms delay before running
			time.Sleep(time.Millisecond * 100)
			drainChan(watcher.Event)
			run(cmd)
		case err := <-watcher.Error:
			log.Fatal("error:", err)
		}
	}
}
