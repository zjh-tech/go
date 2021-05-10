#!/bin/sh

SERVERLIST='sgserver'

ProjectsPath=`pwd`
ProjectBinPath=$ProjectsPath/bin

ulimit -c unlimited

for serv in $SERVERLIST
do 
	cd $ProjectBinPath/$serv 
	nohup  ./$serv > $serv.log 2>&1 &
	sleep 1
	echo "run $serv ok ..."

	echo ""
done 

for serv in $SERVERLIST
do 
ps x | grep $serv | grep -v "vim" | grep -v "redis" | grep -v "grep" | grep -v "rungame"
done 
