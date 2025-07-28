function dump(o)
   if type(o) == 'table' then
      local s = '{ '
      for k,v in pairs(o) do
         if type(k) ~= 'number' then k = '"'..k..'"' end
         s = s .. '['..k..'] = ' .. dump(v) .. ','
      end
      return s .. '} '
   else
      return tostring(o)
   end
end

local cli = shmake.new()

cli:command("mycmd", "My command description",
    function(parsed, command, app)
        print("You activated `mycmd` command")
    end)

cli:run()
