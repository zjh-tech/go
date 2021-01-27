#!/bin/sh

SERVERLIST='registryserver loginserver gatewayserver centerserver hallserver matchserver battleserver dbserver robot'

bin="$GOBIN"

ulimit -c unlimited

for serv in $SERVERLIST
do 
	rm -rf $bin/$serv/core.*
	rm -rf $bin/$serv/log
	rm -rf $bin/$serv/$serv
	rm -rf $bin/$serv/$serv.exe
	rm -rf $bin/$serv/$serv.log
	echo "$serv clear core log exe ok ..."
done 

