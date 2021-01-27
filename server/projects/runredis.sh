#!/bin/sh

#SERVERLIST='6380 6381 6382 6383 6384 6385 6386 6387 6388 6389'
SERVERLIST='6380'

ulimit -c unlimited

for serv in $SERVERLIST
do	
	redis-server /etc/redis$serv.conf	&
	echo "run redis-server /etc/$serv ok ..."
done 

ps x | grep "redis-server" | grep -v "vim" | grep -v "grep" | grep -v "runredis"
