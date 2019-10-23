local shmake = require("shmake.main")
local shell = require("shmake.shell")
local files = require("shmake.files")

function clean()
    files.delete('*.pyc')
    files.delete('file')
    files.delete({ files="some_dir", only_directories=true })
end

function a_script()
    shell.run([[ touch "file" ]])
end

local dev = shmake.register_env{name="dev", default=true}
local prod = shmake.register_env{name="prod"}

shmake.register_task{name="script", fn=a_script, env=dev}
shmake.register_task{name="clean", fn=clean, env=prod}
