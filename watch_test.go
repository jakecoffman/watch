package main

import (
	"testing"
	"time"

	"code.google.com/p/go.exp/fsnotify"
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
	c := make(chan *fsnotify.FileEvent)
	go func() {
		c <- &fsnotify.FileEvent{}
		c <- &fsnotify.FileEvent{}
	}()
	debounce(c, 100*time.Millisecond)
	select {
	case <-c:
		t.Fatal("Debounce did not drain the channel")
	default:
	}
}
