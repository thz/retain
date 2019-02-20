#!/bin/sh
name=retain
build() {
	export GOOS=$1 ; shift
	export GOARCH=$1 ; shift
	go build -o $name-$GOOS-$GOARCH
}

build solaris amd64
build linux arm64
