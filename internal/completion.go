package internal

import (
	"context"
	_ "embed"
	"fmt"
	"strings"

	"github.com/urfave/cli/v3"

	"github.com/mbark/sindr/internal/logger"
)

//go:embed completion/completion.fish
var fishCompletion string

func ConfigureShellCompletionCommand(cmd *cli.Command) {
	action := cmd.Action
	cmd.Action = func(ctx context.Context, command *cli.Command) error {
		args := command.Args().Slice()
		if len(args) > 0 && args[0] == "fish" {
			logger.Print(fishCompletion)
			return nil
		}

		return action(ctx, command)
	}
}

type Completion struct{ Text, Desc string }

// ---- helpers ---------------------------------------------------------------

func flagTakesValue(f cli.Flag) bool {
	switch f.(type) {
	case *cli.BoolFlag:
		return false
	default:
		return true
	}
}

func normalizeFlagSpelling(n string) string {
	// Accept "cache-dir", "--cache-dir", "-f" and normalize to canonical spellings
	if strings.HasPrefix(n, "--") || strings.HasPrefix(n, "-") {
		return n
	}
	if len(n) == 1 {
		return "-" + n
	}
	return "--" + n
}

type ctxState struct {
	curCmd          *cli.Command
	waitingForValue bool
	waitingFlagName string
	helpMode        bool // user typed "help" (or "h") somewhere
}

func walkToCommand(root *cli.Command, tokens []string) (st ctxState) {
	cur := root

	flagLookup := func(cmd *cli.Command) map[string]cli.Flag {
		m := map[string]cli.Flag{}
		for _, f := range cmd.VisibleFlags() {
			for _, nm := range f.Names() {
				m[normalizeFlagSpelling(nm)] = f
			}
		}
		return m
	}

	i := 0
	for i < len(tokens) {
		tok := tokens[i]

		if st.waitingForValue {
			st.waitingForValue = false
			i++
			continue
		}

		// --flag=value
		if strings.HasPrefix(tok, "--") && strings.Contains(tok, "=") {
			name := tok[:strings.IndexByte(tok, '=')]
			if f, ok := flagLookup(cur)[name]; ok && flagTakesValue(f) {
				i++
				continue
			}
		}

		// Flags
		if strings.HasPrefix(tok, "-") {
			if f, ok := flagLookup(cur)[tok]; ok {
				if flagTakesValue(f) {
					st.waitingForValue = true
					st.waitingFlagName = tok
				}
				i++
				continue
			}
			i++
			continue
		}

		// Special-case help tokens: stay at current command, but mark help mode.
		if tok == "help" || tok == "h" {
			st.helpMode = true
			i++
			continue
		}

		// Subcommand descent
		var next *cli.Command
		for _, c := range cur.VisibleCommands() {
			if c.HasName(tok) {
				next = c
				break
			}
		}
		if next != nil {
			cur = next
			i++
			continue
		}

		// Positional arg
		i++
	}

	st.curCmd = cur
	return st
}

// helper: does this command already have a "help" child?
func hasHelpChild(cmd *cli.Command) bool {
	for _, c := range cmd.VisibleCommands() {
		if c.HasName("help") || c.HasName("h") {
			return true
		}
	}
	return false
}

func ComputeCompletions(app *cli.Command, tokens []string, curtok string) []Completion {
	st := walkToCommand(app, tokens)

	// Case 1: currently providing a flag value → no command/flag suggestions
	if st.waitingForValue {
		return nil
	}
	if strings.HasPrefix(curtok, "--") && strings.HasSuffix(curtok, "=") {
		return nil
	}

	// Decide kind of suggestions
	onlyFlags := strings.HasPrefix(curtok, "-")
	// If user is in "help mode", show ONLY subcommands regardless of '-'
	onlySubcommands := st.helpMode || !onlyFlags

	var out []Completion

	if onlySubcommands {
		// Current command's subcommands
		for _, c := range st.curCmd.VisibleCommands() {
			name := c.Name
			if curtok == "" || strings.HasPrefix(name, curtok) {
				out = append(out, Completion{Text: name, Desc: c.Usage})
			}
			for _, a := range c.Aliases {
				if curtok == "" || strings.HasPrefix(a, curtok) {
					out = append(out, Completion{Text: a, Desc: c.Usage + " (alias)"})
				}
			}
		}
		// Synthesize a "help" suggestion if it isn't already present
		if !hasHelpChild(st.curCmd) {
			if curtok == "" || strings.HasPrefix("help", curtok) {
				out = append(out, Completion{
					Text: "help",
					Desc: "Shows a list of commands or help for one command",
				})
			}
			// optional short alias
			if curtok == "" || strings.HasPrefix("h", curtok) {
				out = append(out, Completion{
					Text: "h",
					Desc: "Shows a list of commands or help for one command (alias)",
				})
			}
		}
		return out
	}

	// Otherwise: flags for the current command level
	for _, f := range st.curCmd.VisibleFlags() {
		usage := f.String() // good enough for a description
		for _, n := range f.Names() {
			sp := normalizeFlagSpelling(n)
			if curtok == "" || strings.HasPrefix(sp, curtok) {
				out = append(out, Completion{Text: sp, Desc: usage})
			}
		}
	}
	return out
}

// ---- wire it to the hidden command ----------------------------------------

func CompleteAction(app *cli.Command) cli.ActionFunc {
	return func(ctx context.Context, c *cli.Command) error {
		// Expect: tokens = -opc (previous tokens), curtok = -ct (current token being edited)
		args := c.Args().Slice()
		// The last arg we pass in our fish function is the current token (can be empty).
		var tokens []string
		var curtok string
		if len(args) > 0 {
			tokens = args[:len(args)-1]
			curtok = args[len(args)-1]
		}
		comps := ComputeCompletions(app, tokens, curtok)
		for _, x := range comps {
			if x.Desc != "" {
				logger.Print(fmt.Sprintf("%s\t%s\n", x.Text, x.Desc))
			} else {
				logger.Print(x.Text + "\n")
			}
		}
		return nil
	}
}
