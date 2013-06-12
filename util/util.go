package util

import "fmt"
import "os"

func MaybePanic(err error) {
	if err != nil {
		panic(err)
	}
}

func Debugging() bool {
	return os.Getenv("GOLOG_DEBUG") != ""
}

func Debugf(format string, args... interface{}) {
	if Debugging() {
		fmt.Printf(format, args...)
	}
}
