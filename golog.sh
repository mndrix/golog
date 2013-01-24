#!/bin/sh
go build bin/golog.go &&
rlwrap ./golog $@
