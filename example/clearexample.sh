#!/bin/sh

LIST='config delay logger mysql redis tcpclient tcpserver timer'

path=`pwd`

for example in $LIST
do 
	rm -rf $path/$example/core.*
	rm -rf $path/$example/log
	rm -rf $path/$example/$example
	rm -rf $path/$example/$example.exe
	rm -rf $path/$example/$example.log
	echo "$path/$example clear core log exe ok ..."
done 

