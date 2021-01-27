set serverlist=registryserver;loginserver;gatewayserver;centerserver;hallserver;matchserver;battleserver
set gobin="%GOBIN%"
set projectspath=%cd%
set remainserverlist=%serverlist%
:loop
for /f "tokens=1* delims=;" %%a in ("%remainserverlist%") do (
	set remainserverlist=%%b
	if "%%a" NEQ "" (
	     TASKKILL -F /IM %%a.exe
	)
)
if defined remainserverlist goto :loop
#pause