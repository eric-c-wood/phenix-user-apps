
reg add HKLM\SYSTEM\CurrentControlSet\Control\ComputerName\ActiveComputerName /v ComputerName /t REG_SZ /d ${node['general']['hostname']} /f
reg add HKLM\SYSTEM\CurrentControlSet\Control\ComputerName\ComputerName /v ComputerName /t REG_SZ /d ${node['general']['hostname']} /f
reg add HKLM\SYSTEM\CurrentControlSet\Services\Tcpip\Parameters /v Hostname /t REG_SZ /d ${node['general']['hostname']} /f
reg add HKLM\SYSTEM\CurrentControlSet\Services\Tcpip\Parameters /v "NV Hostname" /t REG_SZ /d ${node['general']['hostname']} /f
shutdown /r /t 0
