
# Configure CLI metadata
shmake.cli(
    name = "shmake example",
    usage = "show how shmake can be used as a library"
)

def useCustomFunction(ctx):
    print(custom_function())

shmake.command(
    name = 'custom',
    help = 'use the custom function',
    action = useCustomFunction,
)
