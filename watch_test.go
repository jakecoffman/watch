package main

import (
	"errors"
	"testing"
	"time"

	"gopkg.in/fsnotify.v1"
)

var parseTests = []struct {
	in  []string
	out []string
}{
	{[]string{"cmd"}, []string{"go", "test"}},
	{[]string{"cmd", "arg"}, []string{"arg"}},
	{[]string{"cmd", "arg", "arg"}, []string{"arg", "arg"}},
}

func Test_parse(t *testing.T) {
	for i, tt := range parseTests {
		out := parse(tt.in)
		for j, val := range out {
			if val != tt.out[j] {
				t.Fatalf("%d. parse(%q) => %q, want %q", i, tt.in, out, tt.out)
			}
		}
	}
}

func Test_debounce(t *testing.T) {
	c := make(chan fsnotify.Event)
	go func() {
		c <- fsnotify.Event{}
		c <- fsnotify.Event{}
	}()
	debounce(c, 100*time.Millisecond)
	select {
	case <-c:
		t.Fatal("Debounce did not drain the channel")
	default:
	}
}

func Test_watcherHandler(t *testing.T) {
	// prepare for mocking
	runTemp := run
	debounceTemp := debounce
	defer func() {
		run = runTemp
		debounce = debounceTemp
	}()

	watcher := &fsnotify.Watcher{
		Events: make(chan fsnotify.Event),
		Errors: make(chan error),
	}

	success := make(chan bool)
	fatal := func(v ...interface{}) { success <- true }
	go func() {
		watcherHandler(watcher, []string{"cmd"}, fatal)
	}()

	// test that when we get an event, the expected functions are called
	run = func(cmd []string) {
		if cmd[0] != "cmd" {
			success <- false
			t.Fatal("Incorrect parameter to run: ", cmd)
		}
		success <- true
	}
	debounce = func(c chan fsnotify.Event, debounceTime time.Duration) {
		success <- true
	}
	watcher.Events <- fsnotify.Event{Name: "TEST"}
	timeout := make(chan bool)

	go func() {
		time.Sleep(1 * time.Second)
	}()

	for i := 0; i < 2; i++ {
		select {
		case s := <-success:
			if s != true {
				t.Fatal("Received failure signal")
			}
		case <-timeout:
			t.Fatal("Timed out while waiting for success signal")
		}
	}

	// watcher errors log
	watcher.Errors <- errors.New("Error!")
	go func() {
		time.Sleep(1 * time.Second)
	}()
	select {
	case s := <-success:
		if s != true {
			t.Fatal("Received failure signal")
		}
	case <-timeout:
		t.Fatal("Timed out while wainting for success signal")
	}
}
