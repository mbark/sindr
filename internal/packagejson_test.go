package internal_test

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/mbark/sindr/internal/sindrtest"
)

func TestSindrLoadPackageJson(t *testing.T) {
	t.Run("loads package.json scripts as commands", func(t *testing.T) {
		run := sindrtest.SetupStarlarkRuntime(t)

		withPackageJson(t, map[string]interface{}{
			"name": "test-project",
			"scripts": map[string]string{
				"build":  "echo Building project",
				"lint":   "echo Running linter",
				"start":  "echo Starting server",
				"deploy": "echo Deploying application",
			},
		})

		sindrtest.WithMainStar(t, `
def test_action(ctx):
    print("Package.json loaded successfully")

cli(name="TestSindrLoadPackageJson")
load_package_json(file="package.json")
command(name="test", action=test_action)
`)
		run()
	})

	t.Run("loads package.json with custom npm binary", func(t *testing.T) {
		run := sindrtest.SetupStarlarkRuntime(t)

		withPackageJson(t, map[string]interface{}{
			"name": "test-project-yarn",
			"scripts": map[string]string{
				"build": "echo Building with yarn",
				"lint":  "echo Linting with yarn",
			},
		})

		sindrtest.WithMainStar(t, `
def test_action(ctx):
    print("Package.json loaded with yarn")

cli(name="TestSindrLoadPackageJson")
load_package_json(file="package.json", bin="yarn")
command(name="test", action=test_action)
`)
		run()
	})

	t.Run("handles empty scripts section", func(t *testing.T) {
		run := sindrtest.SetupStarlarkRuntime(t)

		withPackageJson(t, map[string]interface{}{
			"name":    "empty-scripts-project",
			"scripts": map[string]string{},
		})

		sindrtest.WithMainStar(t, `
def test_action(ctx):
    print("Empty scripts package.json loaded")

cli(name="TestSindrLoadPackageJson")
load_package_json(file="package.json")
command(name="test", action=test_action)
`)
		run()
	})

	t.Run("handles package.json without scripts section", func(t *testing.T) {
		run := sindrtest.SetupStarlarkRuntime(t)

		withPackageJson(t, map[string]interface{}{
			"name":         "no-scripts-project",
			"version":      "1.0.0",
			"description":  "A project without scripts",
			"dependencies": map[string]string{},
		})

		sindrtest.WithMainStar(t, `
def test_action(ctx):
    print("No scripts package.json loaded")

cli(name="TestSindrLoadPackageJson")
load_package_json(file="package.json")
command(name="test", action=test_action)
`)
		run()
	})

	t.Run("handles complex script names", func(t *testing.T) {
		run := sindrtest.SetupStarlarkRuntime(t)

		withPackageJson(t, map[string]interface{}{
			"name": "complex-scripts-project",
			"scripts": map[string]string{
				"build:dev":        "echo Building for development",
				"build:prod":       "echo Building for production",
				"test:unit":        "echo Running unit tests",
				"test:integration": "echo Running integration tests",
				"start:watch":      "echo Starting in watch mode",
			},
		})

		sindrtest.WithMainStar(t, `
def test_action(ctx):
    print("Complex scripts package.json loaded")

cli(name="TestSindrLoadPackageJson")
load_package_json(file="package.json")
command(name="test", action=test_action)
`)
		run()
	})
}

func TestSindrLoadPackageJsonErrors(t *testing.T) {
	t.Run("fails when package.json file does not exist", func(t *testing.T) {
		run := sindrtest.SetupStarlarkRuntime(t)

		sindrtest.WithMainStar(t, `
def test_action(ctx):
    print("Should not reach here but errors are logged")

cli(name="TestSindrLoadPackageJsonErrors")
load_package_json(file="nonexistent.json")
command(name="test", action=test_action)
`)

		run(false)
	})

	t.Run("fails when package.json contains invalid JSON", func(t *testing.T) {
		run := sindrtest.SetupStarlarkRuntime(t)

		writePackageJson(t, []byte(`{
			"name": "invalid-json",
			"scripts": {
				"build": "echo Building"
				// This comment makes it invalid JSON
			}
		}`))

		sindrtest.WithMainStar(t, `
def test_action(ctx):
    print("Should not reach here but errors are logged")

cli(name="TestSindrLoadPackageJsonErrors")
load_package_json(file="invalid.json")
command(name="test", action=test_action)
`)

		run(false)
	})
}

func TestPackageJsonStruct(t *testing.T) {
	t.Run("PackageJson struct unmarshals correctly", func(t *testing.T) {
		run := sindrtest.SetupStarlarkRuntime(t)

		writePackageJson(t, []byte(`{
			"name": "test-project",
			"version": "1.0.0",
			"scripts": {
				"build": "webpack --mode production",
				"test": "jest",
				"start": "webpack serve --mode development",
				"lint": "eslint src/"
			},
			"dependencies": {
				"react": "^18.0.0"
			}
		}`))

		sindrtest.WithMainStar(t, `
def test_action(ctx):
    print("PackageJson struct test completed")

cli(name="TestPackageJsonStruct")
load_package_json(file="package.json")
command(name="test", action=test_action)
`)
		run()
	})
}

func writePackageJson(t *testing.T, jsonData []byte) {
	t.Helper()

	dir, err := os.Getwd()
	require.NoError(t, err)

	packageJsonPath := filepath.Join(dir, "package.json")
	err = os.WriteFile(packageJsonPath, jsonData, 0o644)
	require.NoError(t, err)

	t.Cleanup(func() {
		err := os.Remove(packageJsonPath)
		require.NoError(t, err)
	})
}

func withPackageJson(t *testing.T, data map[string]any) {
	t.Helper()

	jsonData, err := json.Marshal(data)
	require.NoError(t, err)

	writePackageJson(t, jsonData)
}
