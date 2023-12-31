package util

import (
	"fmt"
	"os"
)

func Fprintfln(format string, a ...interface{}) {
	fmt.Fprintf(os.Stderr, format+"\n", a...)
}

func Printfln(format string, a ...interface{}) {
	fmt.Printf(format+"\n", a...)
}
