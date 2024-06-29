
Wmic ComputerSystem where Caption='%ComputerName%' rename '${node['general']['hostname']}'
shutdown /r /t 0
