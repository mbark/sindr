local shmake = require("shmake.main")
local shell = require("shmake.shell")
local files = require("shmake.files")

function prod_clean()
    files.delete('*.pyc')
end

function clean()
    files.delete('*.pyc')
    files.delete('file')
    files.delete({ files="some_dir", only_directories=true })
end

function a_script()
    shell.run([[ touch "file" ]])
end

function start()
    shell.start({
        foo=[[ ping google.com ]],
        bar=[[ ping telness.se ]]
    })
end

local dev = shmake.register_env{name="dev", default=true}
local prod = shmake.register_env{name="prod"}

shmake.register_task{name="script", fn=a_script, env=dev}
shmake.register_task{name="clean", fn=clean, env=dev}
shmake.register_task{name="start", fn=start, env=dev}
shmake.register_task{name="prodclean", fn=prod_clean, env=prod}
