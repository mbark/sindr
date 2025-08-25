cli(
    name = "sindr",
    usage = "✨🔨 Easily configurable project-specific CLI."
)

def test(ctx):
    flags = []
    if ctx.flags.short:
        flags.append('-short')

    shell(string('go test {{.flags}} {{.args}} ./...', flags=' '.join(flags), args=ctx.args.args))

command(
    name = "test",
    help = "run go test",
    action = test,
    args = ['args'],
    flags = {
        "short": {
            "type": "bool",
            "default": True,
            "help": "Use the -short flag when running the tests"
        },
    },
)

