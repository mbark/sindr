
def test_action(ctx):
    result = string('Hello {{.name}}!', name='World')
    if result != 'Hello World!':
        fail('expected "Hello World!", got: ' + str(result))

cli(name="TestTemplateString", usage="Test string templating")
command(name="test", action=test_action)
