package internal

import (
	"fmt"
	"strings"
	"time"

	"github.com/charmbracelet/lipgloss"
)

var timeStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("#666")).Faint(true)

func Log(messages ...string) {
	now := time.Now().Format("15:04:05")
	fmt.Printf("%s %s\n", timeStyle.Render(now), strings.Join(messages, " "))

}
