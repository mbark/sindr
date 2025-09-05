package logger

import (
	"errors"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/x/ansi"
	"go.starlark.net/starlark"
)

var (
	Default         Interface = Logger{}
	DoLogVerbose    bool
	WithLineNumbers bool
	Writer          io.Writer = os.Stdout

	stackStyle        = lipgloss.NewStyle().Foreground(lipgloss.ANSIColor(ansi.Black)).Faint(true)
	errorHeaderStyle  = lipgloss.NewStyle().Foreground(lipgloss.ANSIColor(ansi.Red)).Bold(true)
	errorMessageStyle = lipgloss.NewStyle().Foreground(lipgloss.ANSIColor(ansi.Red)).Padding(0, 2)
)

type Interface interface {
	WithStack(stack starlark.CallStack) Interface

	Print(message string)
	Log(messages ...string)
	LogErr(message string, err error)
	LogVerbose(messages ...string)
}

var _ Interface = Logger{}

type Logger struct {
	stack starlark.CallStack
}

func WithStack(stack starlark.CallStack) Interface {
	return Default.WithStack(stack)
}

func Print(message string) {
	Default.Print(message)
}

func Log(messages ...string) {
	Default.Log(messages...)
}

func LogErr(message string, err error) {
	Default.LogErr(message, err)
}

func LogVerbose(messages ...string) {
	if !DoLogVerbose {
		return
	}

	Default.LogVerbose(messages...)
}

func (l Logger) WithStack(stack starlark.CallStack) Interface {
	l.stack = stack
	return l
}

func (l Logger) Print(message string) {
	_, _ = fmt.Fprint(Writer, message)
}

func (l Logger) Log(messages ...string) {
	if len(l.stack) > 0 && WithLineNumbers {
		_, _ = fmt.Fprintf(Writer, "%s %s\n",
			stackStyle.Render(l.stack[0].Pos.String()),
			strings.Join(messages, " "),
		)
		return
	}

	_, _ = fmt.Fprintf(Writer, "%s\n", strings.Join(messages, " "))
}

func (l Logger) LogErr(message string, err error) {
	l.Log(errorHeaderStyle.Render(message))
	l.Log(errorMessageStyle.Render(err.Error()))

	var serr *starlark.EvalError
	if errors.As(err, &serr) {
		l.Log(errorMessageStyle.Render(serr.CallStack.String()))
	}
}

func (l Logger) LogVerbose(messages ...string) {
	if !DoLogVerbose {
		return
	}

	l.Log(messages...)
}
