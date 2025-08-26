package internal_test

import (
	"testing"

	"github.com/mbark/sindr/internal/sindrtest"
)

func TestDotenv(t *testing.T) {
	t.Run("loads default .env file", func(t *testing.T) {
		sindrtest.Test(t, `
def test_action(ctx):
    # Use exec to create a .env file using Python
    exec('python3', '''
with open('.env', 'w') as f:
    f.write('TEST_VAR=hello\\n')
print('Created .env file')
''')
    
    # Load dotenv
    dotenv()
    
    # Check that the variable is set
    result = shell('echo $TEST_VAR')
    if result.stdout != 'hello':
        fail('expected "hello", got: ' + str(result.stdout))

cli(name="TestDotenv")
command(name="test", action=test_action)
`)
	})

	t.Run("loads multiple env files", func(t *testing.T) {
		sindrtest.Test(t, `
def test_action(ctx):
    # Create multiple env files using Python
    exec('python3', '''
with open('.env.local', 'w') as f:
    f.write('VAR1=value1\\n')
    
with open('.env.production', 'w') as f:
    f.write('VAR2=value2\\n')
    
print('Created env files')
''')
    
    # Load both files
    dotenv(['.env.local', '.env.production'])
    
    # Check both variables are set
    result1 = shell('echo $VAR1')
    if result1.stdout != 'value1':
        fail('expected VAR1="value1", got: ' + str(result1.stdout))
    
    result2 = shell('echo $VAR2')
    if result2.stdout != 'value2':
        fail('expected VAR2="value2", got: ' + str(result2.stdout))

cli(name="TestDotenv")
command(name="test", action=test_action)
`)
	})

	t.Run("skips overriding existing environment variables by default", func(t *testing.T) {
		// Set environment variable at Go level before test
		t.Setenv("EXISTING_VAR", "original")

		sindrtest.Test(t, `
def test_action(ctx):
    # Verify existing variable is set
    result = shell('echo $EXISTING_VAR')
    if result.stdout != 'original':
        fail('initial env var not set correctly, got: ' + str(result.stdout))
    
    # Create .env with conflicting value
    exec('python3', '''
with open('.env', 'w') as f:
    f.write('EXISTING_VAR=new_value\\n')
print('Created .env file')
''')
    
    # Load dotenv without overload
    dotenv()
    
    # Check that original value is preserved
    result = shell('echo $EXISTING_VAR')
    if result.stdout != 'original':
        fail('expected "original" (existing var should not be overridden), got: ' + str(result.stdout))

cli(name="TestDotenv")
command(name="test", action=test_action)
`)
	})

	t.Run("overrides existing variables when overload=True", func(t *testing.T) {
		// Set environment variable at Go level before test
		t.Setenv("OVERRIDE_VAR", "original")

		sindrtest.Test(t, `
def test_action(ctx):
    # Verify existing variable is set
    result = shell('echo $OVERRIDE_VAR')
    if result.stdout != 'original':
        fail('initial env var not set correctly, got: ' + str(result.stdout))
    
    # Create .env file with new value
    exec('python3', '''
with open('.env', 'w') as f:
    f.write('OVERRIDE_VAR=new_value\\n')
print('Created .env file')
''')
    
    # Load dotenv with overload
    dotenv(overload=True)
    
    # Check that variable is overridden
    result = shell('echo $OVERRIDE_VAR')
    if result.stdout != 'new_value':
        fail('expected "new_value", got: ' + str(result.stdout))

cli(name="TestDotenv")
command(name="test", action=test_action)
`)
	})

	t.Run("handles missing env file gracefully", func(t *testing.T) {
		sindrtest.Test(t, `
def test_action(ctx):
    # Try to load non-existent file - should fail
    try:
        dotenv(['nonexistent.env'])
        fail('expected dotenv to fail with non-existent file')
    except:
        pass  # Expected to fail

cli(name="TestDotenv")
command(name="test", action=test_action)
`, sindrtest.ShouldFail())
	})

	t.Run("loads complex env file with different formats", func(t *testing.T) {
		sindrtest.Test(t, `
def test_action(ctx):
    # Create complex .env file using Python
    exec('python3', '''
env_content = """# Comment line
SIMPLE_VAR=simple_value
VAR_WITH_SPACES=value with spaces
QUOTED_VAR="quoted value"
VAR_WITH_EQUALS=key=value
EMPTY_VAR=
"""

with open('.env', 'w') as f:
    f.write(env_content)
print('Created complex .env file')
''')
    
    # Load dotenv
    dotenv()
    
    # Test various formats
    result1 = shell('echo "$SIMPLE_VAR"')
    if result1.stdout != 'simple_value':
        fail('expected SIMPLE_VAR="simple_value", got: ' + str(result1.stdout))
    
    result2 = shell('echo "$VAR_WITH_SPACES"')
    if result2.stdout != 'value with spaces':
        fail('expected VAR_WITH_SPACES="value with spaces", got: ' + str(result2.stdout))
    
    result3 = shell('echo "$QUOTED_VAR"')  
    if result3.stdout != 'quoted value':
        fail('expected QUOTED_VAR="quoted value", got: ' + str(result3.stdout))
    
    result4 = shell('echo "$VAR_WITH_EQUALS"')
    if result4.stdout != 'key=value':
        fail('expected VAR_WITH_EQUALS="key=value", got: ' + str(result4.stdout))
    
    result5 = shell('echo "$EMPTY_VAR"')
    if result5.stdout != '':
        fail('expected EMPTY_VAR to be empty, got: ' + str(result5.stdout))

cli(name="TestDotenv")
command(name="test", action=test_action)
`)
	})

	t.Run("handles env variables with export in shell context", func(t *testing.T) {
		sindrtest.Test(t, `
def test_action(ctx):
    # Create .env file using Python
    exec('python3', '''
with open('.env', 'w') as f:
    f.write('EXPORT_TEST=exported_value\\n')
print('Created .env file')
''')
    
    # Load dotenv
    dotenv()
    
    # Test that variable persists in shell context
    result1 = shell('echo $EXPORT_TEST')
    if result1.stdout != 'exported_value':
        fail('expected "exported_value", got: ' + str(result1.stdout))
    
    # Test in a subshell
    result2 = shell('bash -c "echo $EXPORT_TEST"')
    if result2.stdout != 'exported_value':
        fail('expected variable to be available in subshell, got: ' + str(result2.stdout))

cli(name="TestDotenv")
command(name="test", action=test_action)
`)
	})

	t.Run("multiple dotenv calls work correctly", func(t *testing.T) {
		sindrtest.Test(t, `
def test_action(ctx):
    # Create first .env file using Python
    exec('python3', '''
with open('.env1', 'w') as f:
    f.write('FIRST_VAR=first\\n')
print('Created .env1 file')
''')
    
    # Load first env file
    dotenv(['.env1'])
    
    result1 = shell('echo $FIRST_VAR')
    if result1.stdout != 'first':
        fail('expected FIRST_VAR="first", got: ' + str(result1.stdout))
    
    # Create second .env file using Python
    exec('python3', '''
with open('.env2', 'w') as f:
    f.write('SECOND_VAR=second\\n')
print('Created .env2 file')
''')
    
    # Load second env file 
    dotenv(['.env2'])
    
    result2 = shell('echo $SECOND_VAR')
    if result2.stdout != 'second':
        fail('expected SECOND_VAR="second", got: ' + str(result2.stdout))
    
    # Both variables should be available
    result3 = shell('echo $FIRST_VAR')
    if result3.stdout != 'first':
        fail('expected FIRST_VAR to still be "first", got: ' + str(result3.stdout))

cli(name="TestDotenv")
command(name="test", action=test_action)
`)
	})
}
