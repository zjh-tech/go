set serverlist=tsbalanceserver;tsgateserver;tsrankserver;tsrobot
set gobin="%GOBIN%"
set projectspath=%cd%\..\
set remainserverlist=%serverlist%
:loop
for /f "tokens=1* delims=;" %%a in ("%remainserverlist%") do (
	set remainserverlist=%%b
	if "%%a" NEQ "" (
		cd %projectspath%\rank\%%a
		go build
		move /y "%servername%.exe" %gobin%\rank\%%a
	)
)
if defined remainserverlist goto :loop
pause