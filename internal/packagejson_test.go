package internal_test

import (
	"testing"

	"github.com/mbark/sindr/internal/sindrtest"
)

func TestSindrLoadPackageJson(t *testing.T) {
	t.Run("loads package.json scripts as commands", func(t *testing.T) {
		sindrtest.Test(t, `
def test_action(ctx):
    print("Package.json loaded successfully")

cli(name="TestSindrLoadPackageJson")
load_package_json(file="package.json")
command(name="test", action=test_action)
`, sindrtest.WithPackageJson(map[string]interface{}{
			"name": "test-project",
			"scripts": map[string]string{
				"build":  "echo Building project",
				"lint":   "echo Running linter",
				"start":  "echo Starting server",
				"deploy": "echo Deploying application",
			},
		}))
	})

	t.Run("loads package.json with custom npm binary", func(t *testing.T) {
		sindrtest.Test(t, `
def test_action(ctx):
    print("Package.json loaded with yarn")

cli(name="TestSindrLoadPackageJson")
load_package_json(file="package.json", bin="yarn")
command(name="test", action=test_action)
`, sindrtest.WithPackageJson(map[string]interface{}{
			"name": "test-project-yarn",
			"scripts": map[string]string{
				"build": "echo Building with yarn",
				"lint":  "echo Linting with yarn",
			},
		}))
	})

	t.Run("handles empty scripts section", func(t *testing.T) {
		sindrtest.Test(t, `
def test_action(ctx):
    print("Empty scripts package.json loaded")

cli(name="TestSindrLoadPackageJson")
load_package_json(file="package.json")
command(name="test", action=test_action)
`, sindrtest.WithPackageJson(map[string]interface{}{
			"name":    "empty-scripts-project",
			"scripts": map[string]string{},
		}))
	})

	t.Run("handles package.json without scripts section", func(t *testing.T) {
		sindrtest.Test(t, `
def test_action(ctx):
    print("No scripts package.json loaded")

cli(name="TestSindrLoadPackageJson")
load_package_json(file="package.json")
command(name="test", action=test_action)
`, sindrtest.WithPackageJson(map[string]interface{}{
			"name":         "no-scripts-project",
			"version":      "1.0.0",
			"description":  "A project without scripts",
			"dependencies": map[string]string{},
		}))
	})

	t.Run("handles complex script names", func(t *testing.T) {
		sindrtest.Test(t, `
def test_action(ctx):
    print("Complex scripts package.json loaded")

cli(name="TestSindrLoadPackageJson")
load_package_json(file="package.json")
command(name="test", action=test_action)
`, sindrtest.WithPackageJson(map[string]interface{}{
			"name": "complex-scripts-project",
			"scripts": map[string]string{
				"build:dev":        "echo Building for development",
				"build:prod":       "echo Building for production",
				"test:unit":        "echo Running unit tests",
				"test:integration": "echo Running integration tests",
				"start:watch":      "echo Starting in watch mode",
			},
		}))
	})
}

func TestSindrLoadPackageJsonErrors(t *testing.T) {
	t.Run("fails when package.json file does not exist", func(t *testing.T) {
		sindrtest.Test(t, `
def test_action(ctx):
    print("Should not reach here but errors are logged")

cli(name="TestSindrLoadPackageJsonErrors")
load_package_json(file="nonexistent.json")
command(name="test", action=test_action)
`, sindrtest.ShouldFail())
	})

	t.Run("fails when package.json contains invalid JSON", func(t *testing.T) {
		sindrtest.Test(t, `
def test_action(ctx):
    print("Should not reach here but errors are logged")

cli(name="TestSindrLoadPackageJsonErrors")
load_package_json(file="invalid.json")
command(name="test", action=test_action)
`, sindrtest.WithRawPackageJson(`{
			"name": "invalid-json",
			"scripts": {
				"build": "echo Building"
				// This comment makes it invalid JSON
			}
		}`), sindrtest.ShouldFail())
	})
}

func TestPackageJsonStruct(t *testing.T) {
	t.Run("PackageJson struct unmarshals correctly", func(t *testing.T) {
		sindrtest.Test(t, `
def test_action(ctx):
    print("PackageJson struct test completed")

cli(name="TestPackageJsonStruct")
load_package_json(file="package.json")
command(name="test", action=test_action)
`, sindrtest.WithRawPackageJson(`{
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
	})
}
