package state

import (
	"fmt"
	"io"
	"os"
	"time"

	tea "charm.land/bubbletea/v2"
	"github.com/davecgh/go-spew/spew"
)

var Writer io.Writer

func Log(messages ...string) {
	if Writer == nil {
		return
	}

	_ = TimeStamp()
	spew.Fdump(Writer, messages)
}

func TimeStamp() error {
	_, err := fmt.Fprint(Writer, time.Now().UnixMilli())
	if err != nil {
		return err
	}
	return nil
}

func InitializeLog() {
	if _, ok := os.LookupEnv("DEBUG"); ok {
		file, err := tea.LogToFile("debug.log", "")
		if err != nil {
			// Log to stderr and continue - debug logging failure should not crash the application
			fmt.Fprintf(os.Stderr, "warning: failed to initialize debug log: %v\n", err)
			return
		}
		// Note: We intentionally don't defer close here because Writer needs to remain open
		// for the lifetime of the application. The OS will close it on process exit.
		Writer = file
	}
}
