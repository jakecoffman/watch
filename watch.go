package main

import (
	"fmt"
	"log"
	"os"
	"os/exec"
	"strings"
	"time"
	"gopkg.in/fsnotify.v1"
	"path/filepath"
	"bufio"
	"io"
)

const DEBOUNCE_TIME = 500*time.Millisecond

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

	// attempt to read from .gitignore
	gitignore, err := os.Open(".gitignore")
	var ignores map[string]struct{}
	if err != nil {
		fmt.Println(".gitignore not found")
	} else {
		defer gitignore.Close()
		ignores = getIgnores(gitignore)
	}

	addPathsToWatcher(watcher, ignores)
	watcherHandler(watcher.Events, watcher.Errors, cmd)
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
var debounce = func(c <-chan fsnotify.Event, debounceTime time.Duration) {
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
var watcherHandler = func(events <-chan fsnotify.Event, errors <-chan error, cmd []string) {
	for {
		select {
		case <-events:
			run(cmd)
			debounce(events, DEBOUNCE_TIME)
		case err := <-errors:
			log.Println("Error:", err)
			return
		}
	}
}

func getIgnores(gitignore io.Reader) map[string]struct{} {
	ignores := map[string]struct{}{".git":struct{}{}}
	scanner := bufio.NewScanner(gitignore)
	for scanner.Scan() {
		line := scanner.Text();
		line = strings.TrimSpace(line)
		line = strings.TrimPrefix(line, "/")
		if line != "" && !strings.HasPrefix(line, "#") {
			ignores[line] = struct{}{}
		}
	}
	if err := scanner.Err(); err != nil {
		log.Fatal(err)
	}
	return ignores
}

type watcherLike interface {
	Add(string) error
}

func addPathsToWatcher(watcher watcherLike, ignores map[string]struct{}) error {
	err := watcher.Add(".")
	if err != nil {
		return err
	}
	err = filepath.Walk(".", func(path string, info os.FileInfo, err error) error {
		if err != nil {
			log.Fatal(err)
			return err
		}
		if !info.IsDir() {
			return nil
		}
		for ignorePath, _ := range ignores {
			if strings.HasPrefix(path, ignorePath) {
				return nil
			}
		}
		return watcher.Add(path)
	})
	return err
}
