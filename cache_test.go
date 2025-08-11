package shmake_test

import (
	"testing"
)

func TestDiff(t *testing.T) {
	t.Run("with diff expected", func(t *testing.T) {
		run := setupStarlarkRuntime(t)
		withMainStar(t, `
def test_action(ctx):
	if not cache.diff(name='version', version='1'):
		fail('unexpected diff')

shmake.cli(name="TestDiff", usage="Test diff functionality")
shmake.command(name="test", action=test_action)
`)
		run()
	})

	t.Run("with no diff expected", func(t *testing.T) {
		run := setupStarlarkRuntime(t)
		withMainStar(t, `
def test_action(ctx):
	cache.set_version(name='version', version='1')
	if cache.diff(name='version', version='1'):
		fail('expected no diff')

shmake.cli(name="TestDiff", usage="Test diff functionality")
shmake.command(name="test", action=test_action)
`)
		run()
	})
}

func TestStore(t *testing.T) {
	t.Run("set_version version successfully", func(t *testing.T) {
		run := setupStarlarkRuntime(t)
		withMainStar(t, `
def test_action(ctx):
	cache.set_version(name='test-key', version='v1.0.0')
	
	# Verify it was stored by checking with get_version
	stored = cache.get_version('test-key')
	if stored != 'v1.0.0':
		fail('expected stored version to be v1.0.0, got: ' + str(stored))

shmake.cli(name="TestStore", usage="Test set_version functionality")
shmake.command(name="test", action=test_action)
`)
		run()
	})

	t.Run("set_version with int version", func(t *testing.T) {
		run := setupStarlarkRuntime(t)
		withMainStar(t, `
def test_action(ctx):
	cache.set_version(name='test-int', version=42)
	
	# Verify it was stored by checking with get_version
	stored = cache.get_version('test-int')
	if stored != '42':
		fail('expected stored version to be 42, got: ' + str(stored))

shmake.cli(name="TestStore", usage="Test set_version functionality")
shmake.command(name="test", action=test_action)
`)
		run()
	})
}

func TestWithVersion(t *testing.T) {
	t.Run("runs function when version differs", func(t *testing.T) {
		run := setupStarlarkRuntime(t)
		withMainStar(t, `
def test_action(ctx):
	def version_func():
		print('executing')
		return "executed"
	
	ran = shmake.with_version(version_func, name='test-version', version='v2.0.0')
	
	if not ran:
		fail('expected with_version to return true when function runs')
	
	print('Function executed successfully')

shmake.cli(name="TestWithVersion", usage="Test with_version functionality")
shmake.command(name="test", action=test_action)
`)
		run()
	})

	t.Run("skips function when version matches", func(t *testing.T) {
		run := setupStarlarkRuntime(t)
		withMainStar(t, `
def test_action(ctx):
	def version_func():
		fail('function should not be called when versions match')
	
	# First set_version a version
	cache.set_version(name='skip-test', version='v1.5.0')
	
	# Then try to run with_version with same version
	ran = shmake.with_version(version_func, name='skip-test', version='v1.5.0')
	
	if ran:
		fail('expected with_version to return false when versions match')
	
	print('Version matching test passed - function was correctly skipped')
	
shmake.cli(name="TestWithVersion", usage="Test with_version functionality")
shmake.command(name="test", action=test_action)
`)
		run()
	})

	t.Run("runs function with int version", func(t *testing.T) {
		run := setupStarlarkRuntime(t)
		withMainStar(t, `
def test_action(ctx):
	def version_func():
		print('executing int version test')
		return True

	ran = shmake.with_version(version_func, name='int-version', version=123)
	
	if not ran:
		fail('expected with_version to return true when function runs')
	
	print('Int version function executed successfully')
	
	# Verify version was stored
	stored = cache.get_version('int-version')
	if stored != '123':
		fail('expected stored version to be 123, got: ' + str(stored))

shmake.cli(name="TestWithVersion", usage="Test with_version functionality")
shmake.command(name="test", action=test_action)
`)
		run()
	})
}
