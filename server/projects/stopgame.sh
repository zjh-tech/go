#!/bin/sh

SERVERLIST='registryserver loginserver centerserver gatewayserver hallserver matchserver battleserver dbserver'

for serv in $SERVERLIST
do 
	echo -n "stoping $serv "

	ps ux| grep "$serv" | sed -e '/grep/d'|awk '{print $2}'|xargs kill 2& > /dev/null
	while true 
	do
		echo -n "."
		COUNT=`ps ux |grep "$serv"|grep -v "grep"|grep -v "vim"|wc -l`
		if [ $COUNT -eq 0 ] 
		then
			break 
		fi
		sleep 0.1
	done 

	echo "ok"
done	
ps x | grep "server" | grep -v "vim" | grep -v "redis" | grep -v "grep" | grep -v "stopserver"
