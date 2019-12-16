package main

import (
	"github.com/gen-iot/liblpc/v2"
	"github.com/gen-iot/std"
	"log"
)

func onRemoteStreamRead(sw liblpc.StreamWriter, data []byte, len int) {
	log.Println("read remote conn:", string(data[:len]))
	_ = sw.Close()
}

func onRemoteStreamClose(sw liblpc.StreamWriter, err error) {
	log.Println("remote conn closed,err:", err)
	// it's safe to call Close multi times
	_ = sw.Close() // close remote client
}

func onAccept(ln *liblpc.Listener, newFd int, err error) {
	if err != nil {
		log.Printf("listener got error:%v\n", err)
		return
	}
	stream := liblpc.NewConnStream(ln.Loop().(*liblpc.IOEvtLoop), newFd, onRemoteStreamRead)
	stream.SetOnClose(onRemoteStreamClose)
	stream.Start()
}

func simpleClient(loop *liblpc.IOEvtLoop, addr string) {
	cliFd, err := liblpc.NewConnFd(addr)
	std.AssertError(err, "new client fd failed")
	stream := liblpc.NewConnStream(loop, int(cliFd), nil)
	stream.SetOnConnect(func(sw liblpc.StreamWriter, err error) {
		sw.Write([]byte("hello world!"), true)
	})
	stream.SetOnClose(func(sw liblpc.StreamWriter, err error) {
		log.Println("client close :", err)
		_ = sw.Close() // dont forgot close stream itself
		// break loop...
		loop.Break()
	})
	stream.Start()
}

func main() {
	log.Println("simple listener start...")
	loop, err := liblpc.NewIOEvtLoop(1024 * 4)
	std.AssertError(err, "new event loop")
	defer std.CloseIgnoreErr(loop)
	// create listen fd first!
	listenerFd, err := liblpc.NewListenerFd(
		"127.0.0.1:12345", // serve at
		1024,              // backlog
		true,              // enable reuse addr
		true,              // enable reuse port
	)
	std.AssertError(err, "new listener fd")
	// new listener
	listener := liblpc.NewListener(loop, int(listenerFd), onAccept)
	defer std.CloseIgnoreErr(listener)
	listener.Start()
	// start simple client
	simpleClient(loop, "127.0.0.1:12345")
	//
	loop.Run()
	log.Println("simple listener exit...")
}
