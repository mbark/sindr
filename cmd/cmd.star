def cmd(ctx):
    out = shmake.shell('echo "foobar"', prefix='hej')
    print('output', out)
    print("building", ctx.args.target, 'snake_cased', ctx.flags.some_flag, 'by flag-name', ctx.flags['some-flag'])
