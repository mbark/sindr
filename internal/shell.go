package internal

import (
	"bufio"
	"context"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"os/exec"
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/x/ansi"
	"go.starlark.net/starlark"
)

var (
	commandStyle = lipgloss.NewStyle().
		Foreground(lipgloss.ANSIColor(ansi.Blue)).
		Bold(true)
	stdoutStyle = lipgloss.NewStyle().
		Faint(true).
		Padding(0, 2)
	stderrStyle = lipgloss.NewStyle().
		Foreground(lipgloss.ANSIColor(ansi.Red)).
		Padding(0, 2)
	prefixStyle = lipgloss.NewStyle().
		Foreground(lipgloss.ANSIColor(ansi.BrightBlack)).
		Faint(true)
)

func ShmakeShell(
	thread *starlark.Thread,
	fn *starlark.Builtin,
	args starlark.Tuple,
	kwargs []starlark.Tuple,
) (starlark.Value, error) {
	var command, prefix string
	var noOutput bool
	if err := starlark.UnpackArgs("shell", args, kwargs,
		"command", &command,
		"prefix?", &prefix,
		"no_output?", &noOutput,
	); err != nil {
		return nil, err
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	if prefix != "" {
		Log(prefixStyle.Render(prefix), commandStyle.Render("$ "+command))
	} else {
		Log(commandStyle.Render("$ " + command))
	}

	slog.With(slog.String("command", command)).Debug("running shell command")

	cmd := exec.CommandContext(ctx, "bash", "-c", command) // #nosec G204
	res, err := StartShellCmd(cmd, prefix, noOutput)
	if err != nil {
		return nil, fmt.Errorf("start shell cmd failed: %w", err)
	}

	return res, nil
}

func StartShellCmd(cmd *exec.Cmd, name string, noOutput bool) (*ShellResult, error) {
	stdoutPipe, err := cmd.StdoutPipe()
	if err != nil {
		return nil, fmt.Errorf("stdout pipe: %w", err)
	}

	stderrPipe, err := cmd.StderrPipe()
	if err != nil {
		return nil, fmt.Errorf("stderr pipe: %w", err)
	}

	if err := cmd.Start(); err != nil {
		return nil, fmt.Errorf("cmd start: %w", err)
	}

	var stdoutBuilder, stderrBuilder strings.Builder
	scan := func(pipe io.ReadCloser, builder *strings.Builder, style lipgloss.Style) {
		scanner := bufio.NewScanner(pipe)
		scanner.Split(bufio.ScanLines)
		for scanner.Scan() {
			m := scanner.Text()

			if !noOutput {
				builder.WriteString(m + "\n")
			}
			if name != "" {
				Log(prefixStyle.Render(name), style.Render(m))
			} else {
				Log(style.Render(m))
			}
		}
	}

	scan(stdoutPipe, &stdoutBuilder, stdoutStyle)
	scan(stderrPipe, &stderrBuilder, stderrStyle)
	err = cmd.Wait()
	stdout, stderr := strings.TrimSpace(
		stdoutBuilder.String(),
	), strings.TrimSpace(
		stderrBuilder.String(),
	)
	if err != nil {
		var exitErr *exec.ExitError
		if errors.As(err, &exitErr) {
			return &ShellResult{
				Stdout:   stdout,
				Stderr:   stderr,
				Success:  exitErr.Success(),
				ExitCode: exitErr.ExitCode(),
			}, nil
		}

		return nil, err
	}

	return &ShellResult{
		Stdout:   stdout,
		Stderr:   stderr,
		Success:  cmd.ProcessState.Success(),
		ExitCode: cmd.ProcessState.ExitCode(),
	}, err
}

var (
	_ starlark.Value    = (*ShellResult)(nil)
	_ starlark.HasAttrs = (*ShellResult)(nil)
)

type ShellResult struct {
	Stdout   string
	Stderr   string
	ExitCode int
	Success  bool
}

func (s ShellResult) Attr(name string) (starlark.Value, error) {
	switch name {
	case "stdout":
		return starlark.String(s.Stdout), nil
	case "stderr":
		return starlark.String(s.Stderr), nil
	case "exit_code":
		return starlark.MakeInt(s.ExitCode), nil
	case "success":
		return starlark.Bool(s.Success), nil
	default:
		return nil, nil
	}
}

func (s ShellResult) AttrNames() []string {
	return []string{"stdout", "stderr", "exit_code", "success"}
}

func (s ShellResult) String() string {
	return s.Stdout
}

func (s ShellResult) Type() string {
	return "shell_result"
}

func (s ShellResult) Freeze() {
	// ShellResult is immutable, so no-op
}

func (s ShellResult) Truth() starlark.Bool {
	return starlark.Bool(s.Success)
}

func (s ShellResult) Hash() (uint32, error) {
	return starlark.String(s.Stdout).Hash()
}
