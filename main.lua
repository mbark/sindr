local shmake = require("shmake.main")
local shell = require("shmake.shell")
local files = require("shmake.files")
local cache = require("shmake.cache")
local git = require("shmake.git")
local yarn = require("shmake.yarn")
local run = require("shmake.run")

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
    run.async(shell.run, 'sleep 2; echo "first"')
    run.async(shell.run, 'sleep 2; echo "second"')
    run.await()
    shell.start('echo "fourth"')
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
    head = git.head()
    if cache.diff({ name="tidy", version=head }) then
        shell.run([[ go mod tidy ]])
        cache.store({ name="tidy", version=head })
    end
end

shmake.var({name='Project', value='./../../..'})
shmake.var({name='BackendPath', value='{{.Project}}/backend'})
shmake.var({name='ToolsBin', value='{{.BackendPath}}/backend'})

local dev = shmake.env{name="dev", default=true}
local prod = shmake.env{name="prod"}

shmake.task{name="script", fn=a_script, env=dev}
shmake.task{name="clean", fn=clean, env=dev}
shmake.task{name="start", fn=start, env=dev}

shmake.task{name="mod", fn=mod, env=dev}
shmake.task{name="async", fn=async, env=dev}
shmake.task{name="proto", fn=proto, env=dev}
shmake.task{name="install", fn=install, env=dev}

shmake.task{name="update_mod", fn=update_mod, env=dev, args={foo="bar", bar="{{.Project}}/foo"}}

shmake.task{name="prodclean", fn=prod_clean, env=prod}
