#!/bin/sh

SERVERLIST='registryserver loginserver gatewayserver centerserver hallserver matchserver battleserver dbserver robot'

ProjectsPath=`pwd`
echo ""

bin="$GOBIN"

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
	
	mv -f $serv $GOBIN/$serv/
	echo ""
done 

