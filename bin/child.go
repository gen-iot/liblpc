package main

import (
	"fmt"
	"os"
	"time"
)

func main() {
	for _, line := range os.Environ() {
		fmt.Println(line)
	}
	f := os.NewFile(uintptr(3), "")
	readBuf := make([]byte, 1024)
	for {
		time.Sleep(time.Second)
		_, err := f.WriteString(time.Now().String())
		if err != nil {
			fmt.Println("child write error = ", err)
			break
		}

		n, err := f.Read(readBuf)
		if err != nil {
			fmt.Println("child read error = ", err)
			break
		}
		fmt.Println("child read = ", string(readBuf[:n]))
	}
}
