
# Configure CLI metadata
cli(
    name = "sindr example",
    usage = "show how sindr can be used as a library"
)

def useCustomFunction(ctx):
    print(custom_function())

command(
    name = 'custom',
    help = 'use the custom function',
    action = useCustomFunction,
)
