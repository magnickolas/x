package util

import (
	"fmt"
	"os"

	"runtime/debug"

	"github.com/pkg/errors"
)

type stackTracer interface {
	StackTrace() errors.StackTrace
}

func TrapPanic() {
	if r := recover(); r != nil {
		fmt.Println("stacktrace: \n" + string(debug.Stack()))
		fmt.Println(r)
		os.Exit(1)
	}
}

func Must(err error) {
	if err != nil {
		panic(err)
	}
}
