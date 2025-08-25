package sindr

import (
	"context"
	"fmt"
	"os"
	"path"
	"strings"

	"github.com/spf13/viper"
	"github.com/urfave/cli/v3"
	"go.starlark.net/starlark"
	"go.starlark.net/syntax"

	flag "github.com/spf13/pflag"

	"github.com/mbark/sindr/cache"
	"github.com/mbark/sindr/internal"
	"github.com/mbark/sindr/internal/logger"
	"github.com/mbark/sindr/loader"
)

// StarlarkBuiltin exposes the expected function signature for a starlark builtin function. It's just added here to
// simplify adding additional Globals.
type StarlarkBuiltin = func(thread *starlark.Thread, fn *starlark.Builtin, args starlark.Tuple, kwargs []starlark.Tuple) (starlark.Value, error)

type runOptions struct {
	globals starlark.StringDict
}

var (
	cacheDirKey    = "cache_dir"
	fileNameKey    = "file_name"
	verboseKey     = "verbose"
	noCacheKey     = "no_cache"
	lineNumbersKey = "line_numbers"
)

type RunOption func(o *runOptions, v *viper.Viper)

func WithCacheDir(dir string) RunOption {
	return func(o *runOptions, v *viper.Viper) {
		v.Set(cacheDirKey, dir)
	}
}

func WithFileName(name string) RunOption {
	return func(o *runOptions, v *viper.Viper) {
		v.Set(fileNameKey, name)
	}
}

func WithGlobalValue(name string, value starlark.Value) RunOption {
	return func(o *runOptions, v *viper.Viper) {
		o.globals[name] = value
	}
}

func WithVerboseLogging(verbose bool) RunOption {
	return func(o *runOptions, v *viper.Viper) {
		v.Set(verboseKey, verbose)
	}
}

func WithLineNumbers(lineNumbers bool) RunOption {
	return func(o *runOptions, v *viper.Viper) {
		v.Set(lineNumbersKey, lineNumbers)
	}
}

func WithNoCache(noCache bool) RunOption {
	return func(o *runOptions, v *viper.Viper) {
		v.Set(noCacheKey, noCache)
	}
}

// WithBuiltin does exactly what WithGlobalValue does, but handles the much more common case of wanting to add not just
// any global but specifically a StarlarkBuiltin.
func WithBuiltin(name string, builtin StarlarkBuiltin) RunOption {
	return func(o *runOptions, v *viper.Viper) {
		o.globals[name] = starlark.NewBuiltin(name, builtin)
	}
}

func flagName(s string) string {
	return strings.ReplaceAll(s, "-", "_")
}

func Run(ctx context.Context, args []string, opts ...RunOption) error {
	cacheDir := path.Join(xdgPath("CACHE_HOME", path.Join(os.Getenv("HOME"), ".cache")), "sindr")

	v := viper.New()

	fs := flag.NewFlagSet("sindr", flag.ContinueOnError)

	fs.BoolP(flagName(verboseKey), "v", false, "print logs to stdout")
	fs.BoolP(flagName(noCacheKey), "n", false, "ignore stored values in the cache")
	fs.BoolP(flagName(lineNumbersKey), "l", false, "print logs with Starlark line numbers if possible")
	fs.StringP(flagName(fileNameKey), "f", "sindr.star", "path to the Starlark config file")
	fs.String(flagName(cacheDir), cacheDir, "path to the Starlark config file")
	err := fs.Parse(args)

	if err != nil {
		return fmt.Errorf("parsing sindr flags: %w", err)
	}

	err = v.BindPFlags(fs)
	if err != nil {
		return fmt.Errorf("viper bind flags: %w", err)
	}

	// bind all kebab-cased flags to be accessible via underscore, making "no-cache" available as "no_cache"
	for _, key := range v.AllKeys() {
		v.Set(strings.ReplaceAll(key, "-", "_"), v.Get(key))
	}

	v.SetEnvPrefix("SINDR")
	v.AutomaticEnv()
	v.SetConfigFile("sindr")
	v.AddConfigPath(xdgPath("CONFIG_HOME", path.Join(os.Getenv("HOME"), ".config")))

	options := runOptions{
		globals: starlark.StringDict{},
	}
	for _, o := range opts {
		o(&options, v)
	}

	fmt.Println(v.GetString(cacheDirKey))
	logger.DoLogVerbose = v.GetBool(verboseKey)
	logger.WithLineNumbers = v.GetBool(lineNumbersKey)
	cache.GlobalCache.ForceOutOfDate = v.GetBool(noCacheKey)

	cache.SetCache(v.GetString(cacheDirKey))

	dir, err := findPathUpdwards(v.GetString(fileNameKey))
	if err != nil {
		return err
	}

	err = os.Chdir(dir)
	if err != nil {
		return err
	}

	predeclared := createPredeclaredDict(dir)
	for name, value := range options.globals {
		predeclared[name] = value
	}

	loader.Predeclared = predeclared
	thread := &starlark.Thread{
		Name: "cli",
		Load: loader.Load,
		Print: func(thread *starlark.Thread, msg string) {
			logger.WithStack(thread.CallStack()).Log(msg)
		},
	}
	_, err = starlark.ExecFileOptions(
		&syntax.FileOptions{},
		thread,
		v.GetString(fileNameKey),
		nil,
		predeclared,
	)
	if err != nil {
		return err
	}

	return runCLI(ctx, args)
}

func runCLI(ctx context.Context, args []string) error {
	// TODO: implement it so that all flags are automatically copied over here
	// in order for urfave/cli to not error out on invalid flags, we repeat our globally defined flags above again.
	cliFlags := []cli.Flag{
		&cli.BoolFlag{
			Name:    "verbose",
			Usage:   "print logs to stdout",
			Aliases: []string{"v"},
		},
		&cli.BoolFlag{
			Name:    "no-cache",
			Usage:   "ignore stored values in the cache",
			Aliases: []string{"n"},
		},
		&cli.BoolFlag{
			Name:    "with-line-numbers",
			Usage:   "print logs with Starlark line numbers if possible",
			Aliases: []string{"l"},
		},
	}

	cmd := internal.GlobalCLI.Command.Command
	cmd.Flags = append(cmd.Flags, cliFlags...)

	err := cmd.Run(ctx, args)
	if err != nil {
		return err
	}

	internal.WaitGroup.Wait()
	return nil
}

// createPredeclaredDict creates the predeclared dictionary for Starlark execution.
func createPredeclaredDict(dir string) starlark.StringDict {
	return starlark.StringDict{
		"cli":         starlark.NewBuiltin("cli", internal.SindrCLI),
		"command":     starlark.NewBuiltin("command", internal.SindrCommand),
		"sub_command": starlark.NewBuiltin("sub_command", internal.SindrSubCommand),

		"dotenv": starlark.NewBuiltin("dotenv", internal.SindrDotenv),

		"shell": starlark.NewBuiltin("shell", internal.SindrShell),
		"exec":  starlark.NewBuiltin("exec", internal.SindrExec),

		"string": starlark.NewBuiltin("string", internal.SindrString),

		"start": starlark.NewBuiltin("start", internal.SindrStart),
		"wait":  starlark.NewBuiltin("wait", internal.SindrWait),
		"pool":  starlark.NewBuiltin("pool", internal.SindrPool),

		"newest_ts": starlark.NewBuiltin("newest_ts", internal.SindrNewestTS),
		"oldest_ts": starlark.NewBuiltin("oldest_ts", internal.SindrOldestTS),
		"glob":      starlark.NewBuiltin("glob", internal.SindrGlob),

		"load_package_json": starlark.NewBuiltin(
			"load_package_json",
			internal.SindrLoadPackageJson,
		),
		"cache":       starlark.NewBuiltin("cache", cache.NewCacheValue),
		"current_dir": starlark.String("current_dir"),
	}
}
