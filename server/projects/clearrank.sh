#!/bin/sh

SERVERLIST='tsregistryserver tsbalanceserver tsgateserver tsrankserver tsrobot'

bin="$GOBIN"

ulimit -c unlimited

for serv in $SERVERLIST
do 
	rm -rf $bin/rank/$serv/core.*
	rm -rf $bin/rank/$serv/log
	rm -rf $bin/rank/$serv/$serv
	rm -rf $bin/rank/$serv/$serv.exe
	rm -rf $bin/rank/$serv/$serv.log
	echo "rank/$serv clear core log exe ok ..."
done 

