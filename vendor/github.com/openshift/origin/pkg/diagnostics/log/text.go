package log

import (
	"fmt"
	ct "github.com/daviddengcn/go-colortext"
	"github.com/docker/docker/pkg/term"
	"io"
	"os"
	"strings"
)

type textLogger struct {
	out         io.Writer
	ttyOutput   bool // usually want color; but do not output colors to non-tty
	lastNewline bool // keep track of newline separation
}

func newTextLogger(out io.Writer) *textLogger {
	logger := &textLogger{out: out, lastNewline: true}

	if IsTerminal(out) {
		// only want color sequences to humans, not redirected output (logs, "less", etc.)
		logger.ttyOutput = true
	}
	return logger
}

// cribbed a la "github.com/openshift/origin/pkg/cmd/util"
func IsTerminal(w io.Writer) bool {
	file, ok := w.(*os.File)
	return ok && term.IsTerminal(file.Fd())
}

func (t *textLogger) Write(entry Entry) {
	if t.ttyOutput {
		ct.ChangeColor(entry.Level.Color, entry.Level.Bright, ct.None, false)
	}
	text := strings.TrimSpace(entry.Message)
	if entry.Level.Level >= WarnLevel.Level {
		text = fmt.Sprintf("[%s from %s]\n", entry.ID, entry.Origin) + text
	}
	if strings.Contains(text, "\n") { // separate multiline comments with newlines
		if !t.lastNewline {
			fmt.Fprintln(t.out) // separate from previous one-line log msg
		}
		text = text + "\n"
		t.lastNewline = true
	} else {
		t.lastNewline = false
	}
	fmt.Fprintln(t.out, entry.Level.Prefix+strings.Replace(text, "\n", "\n       ", -1))
	if t.ttyOutput {
		ct.ResetColor()
	}
}
