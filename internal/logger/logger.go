package logger

import (
	"errors"
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/x/ansi"
	"go.starlark.net/starlark"
)

var (
	defaultLogger   = Logger{}
	DoLogVerbose    bool
	WithLineNumbers bool

	stackStyle        = lipgloss.NewStyle().Foreground(lipgloss.ANSIColor(ansi.Black)).Faint(true)
	errorHeaderStyle  = lipgloss.NewStyle().Foreground(lipgloss.ANSIColor(ansi.Red)).Bold(true)
	errorMessageStyle = lipgloss.NewStyle().Foreground(lipgloss.ANSIColor(ansi.Red)).Padding(0, 2)
)

type Logger struct {
	stack starlark.CallStack
}

func WithStack(stack starlark.CallStack) Logger {
	l := defaultLogger
	l.stack = stack
	return l
}

func Log(messages ...string) {
	defaultLogger.Log(messages...)
}

func LogErr(message string, err error) {
	defaultLogger.LogErr(message, err)
}

func LogVerbose(messages ...string) {
	defaultLogger.LogVerbose(messages...)
}

func (l Logger) Log(messages ...string) {
	if len(l.stack) > 0 && WithLineNumbers {
		fmt.Printf("%s %s\n",
			stackStyle.Render(l.stack[0].Pos.String()),
			strings.Join(messages, " "),
		)
		return
	}

	fmt.Printf("%s\n", strings.Join(messages, " "))
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
