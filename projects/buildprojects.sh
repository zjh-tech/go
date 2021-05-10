#!/bin/sh

SERVERLIST='sgserver'

ProjectsPath=`pwd`
ProjectBinPath=$ProjectsPath/bin
echo ""

for serv in $SERVERLIST
do 
	cd $ProjectsPath/$serv

	if [  -n "$1" ]; then
		echo "clean $serv  ..."
		go clean 
	fi
	
	echo "start build $serv  ..."
	#go build
	go build -gcflags '-l -N'
	echo "build $serv ok ..."
	
	mv -f $serv $ProjectBinPath/$serv/
	echo ""
done 

