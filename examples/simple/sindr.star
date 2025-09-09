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
    args = [string_arg('args')],
    flags = [
        bool_flag("short", default=True, usage="Use the -short flag when running the tests")
    ],
)
