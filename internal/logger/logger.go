package logger

import (
	"errors"
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/x/ansi"
	"go.starlark.net/starlark"
)

var DoLogVerbose bool

var (
	errorHeaderStyle  = lipgloss.NewStyle().Foreground(lipgloss.ANSIColor(ansi.Red)).Bold(true)
	errorMessageStyle = lipgloss.NewStyle().Foreground(lipgloss.ANSIColor(ansi.Red)).Padding(0, 2)
)

func Log(messages ...string) {
	fmt.Printf("%s\n", strings.Join(messages, " "))
}

func LogErr(message string, err error) {
	Log(errorHeaderStyle.Render(message))
	Log(errorMessageStyle.Render(err.Error()))

	var serr *starlark.EvalError
	if errors.As(err, &serr) {
		Log(errorMessageStyle.Render(serr.CallStack.String()))
	}
}

func LogVerbose(messages ...string) {
	if !DoLogVerbose {
		return
	}

	fmt.Printf("%s\n", strings.Join(messages, " "))
}
