// +build windows

package fsnotify

import (
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestAttributeChangesIgnored(t *testing.T) {
	watcher := newWatcher(t)
	defer watcher.Close()

	testDir := tempMkdir(t)
	defer os.RemoveAll(testDir)

	// Create a file before watching directory
	testFileAlreadyExists := filepath.Join(testDir, "TestFsnotifyEventsExisting.testfile")
	{
		var f *os.File
		f, err := os.OpenFile(testFileAlreadyExists, os.O_WRONLY|os.O_CREATE, 0666)
		if err != nil {
			t.Fatalf("creating test file failed: %s", err)
		}
		f.Sync()
		f.Close()
	}

	addWatch(t, watcher, testDir)

	// Receive errors on the error channel on a separate goroutine
	go func() {
		for err := range watcher.Errors {
			t.Errorf("error received: %s", err)
		}
	}()

	var eventReceived counter
	go func() {
		for event := range watcher.Events {
			if event.Name == filepath.Clean(testFileAlreadyExists) {
				t.Logf("event received: %s", event)
				eventReceived.increment()
			} else {
				t.Logf("unexpected event received: %s", event)
			}
		}
	}()

	// make the file read-only, which is an attribute change
	err := os.Chmod(testFileAlreadyExists, 0400)
	if err != nil {
		t.Fatalf("Failed to mark file as read-only: %v", err)
	}

	// We expect this event to be received almost immediately, but let's wait 500 ms to be sure
	time.Sleep(500 * time.Millisecond)
	watcher.Close()

	eReceived := eventReceived.value()
	if eReceived != 0 {
		t.Fatalf("should not have received any events, received %d after 500 ms", eReceived)
	}
}
