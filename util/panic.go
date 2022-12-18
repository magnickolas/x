package util

import (
	"fmt"
	"os"

	"github.com/pkg/errors"
)

type stackTracer interface {
	StackTrace() errors.StackTrace
}

func TrapPanic() {
	if r := recover(); r != nil {
		if err, ok := r.(stackTracer); ok {
			for _, f := range err.StackTrace() {
				fmt.Printf("%+s:%d\n", f, f)
			}
		}
		fmt.Println(r)
		os.Exit(1)
	}
}

func Must(err error) {
    if err != nil {
        panic(err)
    }
}
