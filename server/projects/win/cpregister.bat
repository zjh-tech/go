set bin="%GOBIN%"
del /f /s /q %bin%\rank\tsregistryserver\tsregistryserver.exe
del /f /s /q  %bin%\rank\tsregistryserver\registryserver.exe
xcopy  %bin%\registryserver\registryserver.exe  %bin%\rank\tsregistryserver\ /e/y
ren  %bin%\rank\tsregistryserver\registryserver.exe   tsregistryserver.exe
echo " %bin%/rank/tsregistryserver/tsregistryserver ok ..."
pause