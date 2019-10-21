local shmake = require("shmake.main")
local files = require("shmake.files")
local shell = require("shmake.shell")
local yarn = require("shmake.yarn")

function clean()
    files.delete('*.pyc')
    files.delete({ files="__pycache__", only_directories=true })
    files.delete({ files="*.*", dir="triresolve/media" })
    files.delete("triresolve/media_source/build")
    files.delete("triresolve/media_source/elm/elm-stuff")
    files.delete("triresolve/media_source/elm/.elm")

function prod_assets()
    shell.run([[
        rm -rf triresolve/media_source/build/
        mkdir -p triresolve/media_source/build/
        yarn run build:prod
        cp -a node_modules/@trioptima/trids/styles/{fonts,images} triresolve/media_source/build/
        cp -a node_modules/@trioptima/trids/dist/favicon.ico triresolve/media_source/build/images/
        cp -a triresolve/media_source/static/* triresolve/media_source/build/
        $(PYTHON_) triresolve/manage.py collectstatic --noinput | tail -1
    ]])

function dev_assets()
    yarn.install()
    yarn.run('gulp build')

    files_to_cp = { { from: 'node_modules/@trioptima/trids/styles/{fonts,images}' to:'triresolve/media/' },
        { from: 'node_modules/@trioptima/trids/dist/favicon.ico', to: 'triresolve/media/images/' },
        { from: 'triresolve/media_source/static/*', to: 'triresolve/media/' } }

    for _, operation in files_to_cp do
        files.copy(operation)
    end

function assets_live()
    yarn.run('gulp dev')

local dev = m.add_env{name="dev", default=true}
local prod = m.add_env{name="prod"}

shmake.add_task{fn=clean, name="clean"}
shmake.add_task{fn=prod_assets, env=prod, name="assets"}
shmake.add_task{fn=dev_assets, env=dev, name="assets"}
