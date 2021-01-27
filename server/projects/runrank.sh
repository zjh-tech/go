#!/bin/sh

SERVERLIST='tsregistryserver tsbalanceserver tsgateserver tsrankserver'

bin="$GOBIN"

ulimit -c unlimited


for serv in $SERVERLIST
do 
	cd $bin/rank/$serv 
	nohup  ./$serv > $serv.log 2>&1 &
	echo "run rank/$serv ok ..."
	echo ""
done 

ps x | grep "server" | grep -v "vim" | grep -v "redis" | grep -v "grep" | grep -v "runrank"
