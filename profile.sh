#!/bin/sh
go test -c
./golog.test -test.run=none -test.bench=$2 -test.$1profile=$1.profile
