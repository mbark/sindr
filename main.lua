local shmake = require("shmake.main")
local shell = require("shmake.shell")
local files = require("shmake.files")
local cache = require("shmake.cache")
local git = require("shmake.git")
local yarn = require("shmake.yarn")
local run = require("shmake.run")

shmake.var({name='Project', value='./../../..'})
shmake.var({name='BackendPath', value='{{.Project}}/backend'})
shmake.var({name='ToolsBin', value='{{.BackendPath}}/backend'})

local dev = shmake.env{name="dev", default=true}
local prod = shmake.env{name="prod"}

function prod_clean()
    files.delete('*.pyc')
end

function clean(args)
    files.delete('file')
    files.delete('file2')
    files.delete({ files="some_dir", only_directories=true })
    files.delete({ files="nested", only_directories=true })
end

function clean_watch()
    run.watch({
        file={fn=files.delete, args='file2', watch='./file3'}
    })
end


function install()
    files.chdir('./examples/yarn')
    yarn.install()
    yarn.run('prettier -w package.json')
    files.popdir()
end

function async()
    run.async(shell.run, 'sleep 2; echo "{{.Project}}"; echo "first"')
    run.async(shell.run, 'sleep 2; echo "second"')
    run.await()
    shell.run('echo "fourth"')
end

function a_script()
    shell.run([[ touch "file" ]])
    files.copy({ from='file', to='file2' })
    files.mkdir({dir='nested/directory', all=true})
end

function start()
    shell.start({
        foo={cmd=[[ ping google.com ]], watch="./file"},
        bar={cmd=[[ ping telness.se ]], watch="./file2"}
    })
end

function mod()
    gomod_modtime = tostring(files.newest_ts("go.mod"))
    if cache.diff({ name="go.mod", version=gomod_modtime }) then
        shell.run([[ go mod tidy ]])
        cache.store({ name="go.mod", version=gomod_modtime })
    end
end

function proto()
    if files.newest_ts('*.proto', '{{.BACKEND_PATH}}/bin/inject') > files.oldest_ts('*.pb.go') then
        shell.run([[
            go mod vendor
            {{.PROTOC}} \
                -I {{.PROTO_INC}} \
                --validate_out="lang=go:.." \
                --twirp_out=.. \
                --go_out=.. \
                {{.RPC_V1_SRC}}
            {{.PROTOC}} \
                -I {{.PROTO_INC}} \
                --validate_out="lang=go:.." \
                --twirp_out=.. \
                --go_out=.. \
                {{.RPC_V2_SRC}}
            {{.PROTOC}} \
                -I {{.PROTO_INC}} \
                --validate_out="lang=go:.." \
                --go_out=.. \
                --gotemplate_out=all=true:. \
                {{.PROTO_DIR}}/telness/event/v1/*.proto
            {{.TOOLS_BIN}}/inject {{.BACKEND_PATH}}/src/telness/internal/rpc/*.pb.go
        ]])
        files.delete('{{.PROTO_VND}}')
    end
end

function update_mod(args)
    -- local variables can be accessed CamelCased in templates
    shell.run([[ echo '{{.Bar}}' ]])
    print(args['bar'])
    head = git.head()
    if cache.diff({ name="tidy", version=head }) then
        shell.run([[ go mod tidy ]])
        cache.store({ name="tidy", version=head })
    end
end

shmake.cmd{name="script", fn=a_script, env=dev}
shmake.cmd{name="clean", fn=clean, env=dev}
shmake.cmd{name="start", fn=start, env=dev}

shmake.cmd{name="mod", fn=mod, env=dev}
shmake.cmd{name="async", fn=async, env=dev}
shmake.cmd{name="proto", fn=proto, env=dev}
shmake.cmd{name="install", fn=install, env=dev}

shmake.cmd{name="update_mod", fn=update_mod, env=dev, args={foo="bar", bar="{{.Project}}/foo"}}

shmake.cmd{name="prodclean", fn=prod_clean, env=prod}
