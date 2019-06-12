package utils

import (
	"io"
)

func PanicIfError(err error) {
	if err != nil {
		panic(err)
	}
}

func Assert(b bool, msg string) {
	if !b {
		panic(msg)
	}
}

func CloseIgnoreErr(clo io.Closer) {
	_ = clo.Close()
}
