#!/bin/sh

SERVERLIST='sgserver'

ProjectsPath=`pwd`
ProjectBinPath=$ProjectsPath/bin

ulimit -c unlimited

for serv in $SERVERLIST
do 
	rm -rf $serv/log
	rm -rf $ProjectBinPath/$serv/core.*
	rm -rf $ProjectBinPath/$serv/log
	rm -rf $ProjectBinPath/$serv/$serv
	rm -rf $ProjectBinPath/$serv/$serv.exe
	rm -rf $ProjectBinPath/$serv/$serv.log
	echo "$serv clear core log exe ok ..."
done 

