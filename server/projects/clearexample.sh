#!/bin/sh

LIST='config logger mysql redis tcpclient tcpserver timer'

for example in $LIST
do 
	rm -rf example/$example/log
	rm -rf example/$example/$example
	echo "$example clear  log exe ok ..."
done 

