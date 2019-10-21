local shmake = require("shmake.main")
local shell = require("shmake.shell")

function a_script()
    shell.run([[ touch "file" ]])
end

shmake.register_task{name="script", fn=a_script}
