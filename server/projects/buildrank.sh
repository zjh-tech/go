#!/bin/sh

SERVERLIST='tsbalanceserver tsgateserver tsrankserver tsrobot'

ProjectsPath=`pwd`
echo ""

bin="$GOBIN"

for serv in $SERVERLIST
do 
	cd $ProjectsPath/rank/$serv
	if [  -n "$1" ]; then
		echo "clean rank/$serv  ..."
		go clean 
	fi
	
	echo "start build rank/$serv  ..."
	#go build
	go build -gcflags '-l -N'
	echo "build rank/$serv ok ..."
	
	mv -f $serv $GOBIN/rank/$serv
	echo ""
done 

