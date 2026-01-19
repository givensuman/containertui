package context

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
		var err error
		file, err := tea.LogToFile("debug.log", "")
		if err != nil {
			panic(err)
		}
		defer func() {
			err = file.Close()
		}()
		if err != nil {
			panic(err)
		}

		Writer = file
	}
}
