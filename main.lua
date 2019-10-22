local shmake = require("shmake.main")
local shell = require("shmake.shell")
local files = require("shmake.files")

function clean()
    files.delete('*.pyc')
    files.delete({ files="some_dir", only_directories=true })
end

function a_script()
    shell.run([[ touch "file" ]])
end

shmake.register_task{name="script", fn=a_script}
shmake.register_task{name="clean", fn=clean}
