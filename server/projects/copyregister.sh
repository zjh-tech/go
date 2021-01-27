#!/bin/sh


bin="$GOBIN"

rm -rf $bin/rank/tsregistryserver/tsregistryserver
rm -rf $bin/rank/tsregistryserver/registryserver
cp -r $bin/registryserver/registryserver $bin/rank/tsregistryserver/
mv $bin/rank/tsregistryserver/registryserver $bin/rank/tsregistryserver/tsregistryserver
echo "$bin/rank/tsregistryserver/tsregistryserver ok ..."

