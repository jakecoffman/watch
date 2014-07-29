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
	cmd := parse(os.Args)
	log.Printf("Running %v", strings.Join(cmd, " "))
	log.Println("Press Ctl-C to stop watching\n")

	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		log.Fatal(err)
	}

	// run once when first starting up.
	run(cmd)

	// kick off the goroutine that runs the command when the watcher fires
	go watcherHandler(watcher, cmd, log.Fatal)

	// watches cwd, could be configurable
	err = watcher.Watch(".")
	if err != nil {
		log.Fatal(err)
	}

	// block forever-ish
	<-make(chan bool)
}

// parses command line arguments (rudimentary)
var parse = func(args []string) []string {
	var cmd []string
	if len(args) == 1 {
		cmd = []string{"go", "test"}
	} else {
		cmd = args[1:]
	}
	return cmd
}

// do a run of the command
var run = func(cmd []string) {
	log.Println(cmd, "************************************************")

	c := exec.Command(cmd[0], cmd[1:]...)
	out, err := c.CombinedOutput()
	fmt.Print(string(out))

	if err != nil {
		log.Printf("%v\n\n", err)
	} else {
		log.Println("Run complete\n")
	}
}

// takes a channel and reads everything from it for a given amount of time
var debounce = func(c chan *fsnotify.FileEvent, debounceTime time.Duration) {
	timeout := make(chan bool)
	defer func() { close(timeout) }()
	go func() {
		time.Sleep(debounceTime)
		timeout <- true
	}()
	for {
		select {
		case <-c:
			// do nothing
		case <-timeout:
			return
		}
	}
}

// handles watcher events by running the given command
var watcherHandler = func(watcher *fsnotify.Watcher, cmd []string, fatal func(v ...interface{})) {
	for {
		select {
		case <-watcher.Event:
			run(cmd)
			debounce(watcher.Event, 100*time.Millisecond)
		case err := <-watcher.Error:
			fatal("error:", err)
		}
	}
}
