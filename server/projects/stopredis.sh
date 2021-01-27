#!/bin/sh

#SERVERLIST='6380 6381 6382 6383 6384 6385 6386 6387 6388 6389'
SERVERLIST='6380'

for serv in $SERVERLIST
do 
	echo -n "stoping redis-server $serv"

	ps ux| grep "redis-server" | grep "$serv" | sed -e '/grep/d'|awk '{print $2}'|xargs kill 2& > /dev/null
	while true 
	do
		echo -n "."
		COUNT=`ps ux | grep "redis-server" |grep "$serv"|grep -v "grep"|grep -v "vim"|wc -l`
		if [ $COUNT -eq 0 ] 
		then
			break 
		fi
		sleep 0.1
	done 

	echo "ok"
done	
ps x | grep "redis-server" | grep -v "vim" | grep -v "grep" | grep -v "stopredis"
