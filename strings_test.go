package shmake_test

import (
	"testing"

	"github.com/mbark/shmake"
)

func TestTemplateString(t *testing.T) {
	t.Run("renders simple template", func(t *testing.T) {
		r := setupRuntime(t)
		withMainLua(t, `
local shmake = require("shmake.main")

local cli = shmake.command('TestTemplateString')

cli:command('test'):action(function()
	local result = shmake.string('Hello {{.name}}!', {name = 'World'})
	if result ~= 'Hello World!' then
		error('expected "Hello World!", got: ' .. tostring(result))
	end
end)

cli:run()
`)
		shmake.RunWithRuntime(t.Context(), r)
	})

	t.Run("uses global variables", func(t *testing.T) {
		r := setupRuntime(t)
		withMainLua(t, `
local shmake = require("shmake.main")

version = "1.2.3"
name = "myapp"

local cli = shmake.command('TestTemplateString')

cli:command('test'):action(function()
	local result = shmake.string('{{.name}} version {{.version}}')
	if result ~= 'myapp version 1.2.3' then
		error('expected "myapp version 1.2.3", got: ' .. tostring(result))
	end
end)

cli:run()
`)
		shmake.RunWithRuntime(t.Context(), r)
	})

	t.Run("globals override local context", func(t *testing.T) {
		r := setupRuntime(t)
		withMainLua(t, `
local shmake = require("shmake.main")

name = "global"

local cli = shmake.command('TestTemplateString')

cli:command('test'):action(function()
	local result = shmake.string('Hello {{.name}}!', {name = 'Local'})
	if result ~= 'Hello global!' then
		error('expected "Hello global!", got: ' .. tostring(result))
	end
end)

cli:run()
`)
		shmake.RunWithRuntime(t.Context(), r)
	})

	t.Run("handles different variable types", func(t *testing.T) {
		r := setupRuntime(t)
		withMainLua(t, `
local shmake = require("shmake.main")

str_var = "text"
num_var = 42
bool_var = true

local cli = shmake.command('TestTemplateString')

cli:command('test'):action(function()
	local result = shmake.string('{{.str_var}} {{.num_var}} {{.bool_var}}')
	if result ~= 'text 42 true' then
		error('expected "text 42 true", got: ' .. tostring(result))
	end
end)

cli:run()
`)
		shmake.RunWithRuntime(t.Context(), r)
	})

	t.Run("renders template without additional context", func(t *testing.T) {
		r := setupRuntime(t)
		withMainLua(t, `
local shmake = require("shmake.main")

greeting = "Hello"
target = "World"

local cli = shmake.command('TestTemplateString')

cli:command('test'):action(function()
	local result = shmake.string('{{.greeting}} {{.target}}!')
	if result ~= 'Hello World!' then
		error('expected "Hello World!", got: ' .. tostring(result))
	end
end)

cli:run()
`)
		shmake.RunWithRuntime(t.Context(), r)
	})

	t.Run("handles template with conditional logic", func(t *testing.T) {
		r := setupRuntime(t)
		withMainLua(t, `
local shmake = require("shmake.main")

debug = true

local cli = shmake.command('TestTemplateString')

cli:command('test'):action(function()
	local result = shmake.string('{{if .debug}}DEBUG: {{end}}message', {message = 'test'})
	if result ~= 'DEBUG: message' then
		error('expected "DEBUG: message", got: ' .. tostring(result))
	end
end)

cli:run()
`)
		shmake.RunWithRuntime(t.Context(), r)
	})

	t.Run("handles template with range", func(t *testing.T) {
		r := setupRuntime(t)
		withMainLua(t, `
local shmake = require("shmake.main")

local cli = shmake.command('TestTemplateString')

cli:command('test'):action(function()
	local items = {'apple', 'banana', 'cherry'}
	local result = shmake.string('{{range .items}}{{.}} {{end}}', {items = items})
	if result ~= 'apple banana cherry ' then
		error('expected "apple banana cherry ", got: ' .. tostring(result))
	end
end)

cli:run()
`)
		shmake.RunWithRuntime(t.Context(), r)
	})

	t.Run("handles complex nested templates", func(t *testing.T) {
		r := setupRuntime(t)
		withMainLua(t, `
local shmake = require("shmake.main")

app_name = "myapp"
version = "1.0.0"

local cli = shmake.command('TestTemplateString')

cli:command('test'):action(function()
	local config = {
		env = "production",
		port = 8080,
		features = {'auth', 'logging', 'metrics'}
	}
	
	local template = [[
{{.app_name}} v{{.version}}
Environment: {{.config.env}}
Port: {{.config.port}}
Features:{{range .config.features}} {{.}}{{end}}
]]
	
	local result = shmake.string(template, {config = config})
	local expected = [[
myapp v1.0.0
Environment: production
Port: 8080
Features: auth logging metrics
]]
	
	if result ~= expected then
		error('template did not render correctly')
	end
end)

cli:run()
`)
		shmake.RunWithRuntime(t.Context(), r)
	})
}
