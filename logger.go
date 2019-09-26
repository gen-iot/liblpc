package liblpc

import (
	"log"
)

var Debug = false

func stdLog(v ...interface{}) {
	if !Debug {
		return
	}
	log.Println(v...)
}

func stdLogf(format string, v ...interface{}) {
	if !Debug {
		return
	}
	log.Printf("liblpc:"+format, v...)
}
