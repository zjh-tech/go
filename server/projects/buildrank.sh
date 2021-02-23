#!/bin/sh

SERVERLIST='rankserver'

ProjectsPath=`pwd`
echo ""

bin="$GOBIN"

for serv in $SERVERLIST
do 
	cd $ProjectsPath/rankserver
	if [  -n "$1" ]; then
		echo "clean rankserver  ..."
		go clean 
	fi
	
	echo "start build rankserver  ..."
	#go build
	go build -gcflags '-l -N'
	echo "build rankserver ok ..."
	
	mv -f $serv $GOBIN/rankserver
	echo ""
done 

