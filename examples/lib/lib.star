
# Configure CLI metadata
cli(
    name = "shmake example",
    usage = "show how shmake can be used as a library"
)

def useCustomFunction(ctx):
    print(custom_function())

command(
    name = 'custom',
    help = 'use the custom function',
    action = useCustomFunction,
)
