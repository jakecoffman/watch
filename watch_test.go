package main

import (
	"testing"
	"time"
	"gopkg.in/fsnotify.v1"
	"errors"
	"bytes"
	"reflect"
	"os/exec"
)

var parseTests = []struct {
	in  []string
	out []string
}{
	{[]string{"cmd"}, []string{"go", "test"}},
	{[]string{"cmd", "arg"}, []string{"arg"}},
	{[]string{"cmd", "arg", "arg"}, []string{"arg", "arg"}},
}

func Test_Integration(t *testing.T) {
	var c *exec.Cmd
	go func() {
		c = exec.Command("go", "run", "watch.go", "touch", "integrationtest")
		_, err := c.CombinedOutput()
		if err != nil {
			t.Fatal(err)
		}
	}()
	time.Sleep(1*time.Second)
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
	// mock run and debounce
	mockRunCalled := false
	mockDebounceCalled := false
	mockRun := func(cmd []string) {
		mockRunCalled = true
	}
	mockDebounce := func(c <-chan fsnotify.Event, debounceTime time.Duration) {
		mockDebounceCalled = true
	}
	runBackup := run
	debounceBackup := debounce
	defer func(){
		run = runBackup
		debounce = debounceBackup
	}()
	run = mockRun
	debounce = mockDebounce

	errorChannel := make(chan error)
	eventChannel := make(chan fsnotify.Event)
	defer func() {
		close(errorChannel)
		close(eventChannel)
	}()

	semaphore := make(chan struct{})
	go func() {
		watcherHandler(eventChannel, errorChannel, []string{"command"})
		close(semaphore)
	}()

	errorChannel <- errors.New("not a real error")
	<-semaphore

	if mockRunCalled != false || mockDebounceCalled != false {
		t.Error("error did not have the desired effect:", mockRunCalled, mockDebounceCalled)
	}

	hit := false
	go func() {
		watcherHandler(eventChannel, errorChannel, []string{"command"})
		hit = true
	}()

	eventChannel <- fsnotify.Event{}
	time.Sleep(1*time.Second)

	if hit != false || mockRunCalled != true || mockDebounceCalled != true {
		t.Error("event did not have the desired effect:", mockRunCalled, mockDebounceCalled)
	}
}

func Test_getIgnores(t *testing.T) {
	test := `#comment
/thing
hello

#commend
hi
`
	expected := map[string]struct{}{
		".git": struct{}{},
		"thing": struct{}{},
		"hello": struct{}{},
		"hi": struct{}{},
	}
	ignores := getIgnores(bytes.NewBufferString(test))
	if !reflect.DeepEqual(expected, ignores) {
		t.Error("not equal:", ignores)
	}
}