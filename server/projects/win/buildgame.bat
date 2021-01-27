set serverlist=registryserver;loginserver;gatewayserver;centerserver;hallserver;matchserver;battleserver;robot
set gobin="%GOBIN%"
set projectspath=%cd%\..\
set remainserverlist=%serverlist%
:loop
for /f "tokens=1* delims=;" %%a in ("%remainserverlist%") do (
	set remainserverlist=%%b
	if "%%a" NEQ "" (
		cd %projectspath%/%%a
		go build
		move /y "%%a.exe" %gobin%\%%a
	)
)
if defined remainserverlist goto :loop
#pause