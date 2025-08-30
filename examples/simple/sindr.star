cli(
    name = "sindr",
    usage = "✨🔨 Project-specific commands as a CLI. "
)

def test(ctx):
    shell('go test {{.flags}} {{.args}} ./...',
        flags='-short' if ctx.short else '')

command(
    name = "test",
    usage = "run go test",
    action = test,
    args = ['args'],
    flags = [
        {
            "name": "short",
            "type": "bool",
            "default": True,
            "usage": "Use the -short flag when running the tests"
        },
    ],
)

