#!/bin/bash

VERSION=`make version`
DIR=./release/
mkdir -p $DIR
for ARCH in 386 amd64 arm arm64
do    
    GOOS=linux GOARCH=$ARCH make BUILD_DIR=$DIR build
    tar -czf $DIR/vpnlist.${VERSION}.${ARCH}.tgz -C $DIR vpnlist
    rm $DIR/vpnlist
done