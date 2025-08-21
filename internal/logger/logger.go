package logger

import (
	"fmt"
	"strings"
	"time"

	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/x/ansi"
)

var DoLogVerbose bool

var (
	timeStyle         = lipgloss.NewStyle().Foreground(lipgloss.Color("#666")).Faint(true)
	errorHeaderStyle  = lipgloss.NewStyle().Foreground(lipgloss.ANSIColor(ansi.Red)).Bold(true)
	errorMessageStyle = lipgloss.NewStyle().Foreground(lipgloss.ANSIColor(ansi.Red)).Padding(0, 2)
)

func Log(messages ...string) {
	now := time.Now().Format("15:04:05")
	fmt.Printf("%s %s\n", timeStyle.Render(now), strings.Join(messages, " "))
}

func LogErr(message string, err error) {
	Log(errorHeaderStyle.Render("async function failed"))
	Log(errorMessageStyle.Render(err.Error()))
}

func LogVerbose(messages ...string) {
	if !DoLogVerbose {
		return
	}

	now := time.Now().Format("15:04:05")
	fmt.Printf("%s %s\n", timeStyle.Render(now), strings.Join(messages, " "))
}
