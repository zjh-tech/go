#!/bin/sh

MODULELIST='elog edb'

ProjectsPath=`pwd`
echo ""

for module in $MODULELIST
do 		
	rm -rf engine/$module/log					
	echo "$module clear core log exe ok ..."
done 

