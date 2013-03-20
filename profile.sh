#!/bin/sh
go test -c
./golog.test -test.run=none -test.bench=. -test.cpuprofile=cpu.profile -test.memprofile=mem.profile
