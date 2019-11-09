# liblpc

High performance async network io library

![go report](https://goreportcard.com/badge/github.com/gen-iot/liblpc)
![license](https://img.shields.io/badge/license-MIT-brightgreen.svg)

## First
if you have **any** good feature suggestions or bug fixed ,
 **any** [Pull Request](https://github.com/gen-iot/liblpc/pulls) or [Issues](https://github.com/gen-iot/liblpc/issues) are **welcome**!

## Overview

- World Based On **Event Loop** 😎
- UnBuffered/Buffered Stream 🚀
- Timers: IO Timeout, Delay... ⏰
- DNS Resolve 🌐
- Lots Of Unix Socket API Toolbox 🔧
- Thread(*Goroutine*) Safe! 👍


## Getting Started

### Create Pure EventLoop:

```go
loop,err := liblpc.NewEventLoop()
std.AssertError(err, "new pure event loop")
```

### Create IO EventLoop:

```go
loop, err := liblpc.NewIOEvtLoop(1024 * 4)
std.AssertError(err, "new io event loop")
```

### Loop Lifecycle:

**exit a loop**
```go
// just call loop.Break in anywhere
loop.Break()
```
📌`Loop.'Close' can't stop a loop but Loop.'Break' can.`

📌`Loop.'Close' use to cleanup a loop`

**Cleanup a loop**

```go
loop ,err := liblpc.NewEventLoop()
std.AssertError(err, "new event loop")
defer loop.Close()
```

**Run loop synchronously**
```go
// block until break loop called
loop.Run()
```

**Run loop asynchronously😂😂😂**
```go
go func(){ loop.Run() }()
```

### Create Listener:

```go
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
listener.Start()
```

### Accept New Conn Stream:

```go
// 📌note: in accept callback
stream := liblpc.NewConnStream(
  ln.Loop().(*liblpc.IOEvtLoop), // cast Loop to IOEventLoop 
  newFd,                         // incoming fd
  onStreamRead,                  // read callback
  )
stream.SetOnClose(onStreamClose) // register close callback
stream.Start()
```

### Create Client Stream:

```go
cliFd, err := liblpc.NewConnFd(addr)
std.AssertError(err, "new client fd failed")
stream := liblpc.NewConnStream(loop, int(cliFd), nil)
stream.SetOnConnect(func(sw liblpc.StreamWriter, err error) {
  sw.Write([]byte("hello world!"), true)
})
stream.SetOnClose(func(sw liblpc.StreamWriter, err error) {
  log.Println("client close :", err)
  // break loop...
  loop.Break()
})
stream.Start()
```
📌`Stream.'Close' is safe to invoke multi times`

📌`Anytime you can't find out whether if Stream is 'Closing' or really been 'Closed',Just invoke  Stream.'Close'`


### Example: Simple **Read/Write/Close** 

```go

package main

import (
	"github.com/gen-iot/liblpc"
	"github.com/gen-iot/std"
	"log"
)

func onStreamRead(sw liblpc.StreamWriter, data []byte, len int) {
  // print client data in string format
	log.Println("on read:", string(data[:len]))
	_ = sw.Close()
}

func onStreamClose(sw liblpc.StreamWriter, err error) {
	log.Println("conn closed,err:", err)
	_ = sw.Close() // close remote client
}

func onAccept(ln *liblpc.Listener, newFd int, err error) {
	if err != nil {
		log.Printf("listener got error:%v\n", err)
		return
	}
	stream := liblpc.NewConnStream(
    ln.Loop().(*liblpc.IOEvtLoop), // cast Loop to IOEventLoop 
    newFd,                         // incoming fd
    onStreamRead,                  // read callback
    )
	stream.SetOnClose(onStreamClose) // register close callback
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
    // close itself
    _ = sw.Close()
		// break loop...
		loop.Break()
	})
	stream.Start()
}

func main() {
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
}
```

## More EventLoop Backend

|Platform|Backend| Support|
|:--:|:--:|:--:|
| Linux | Epoll | Fully Support 😎 |
| OS X |  Kqueue | Not yet 🥺 | 
| Windows | IOCP | Not yet 🥺|
| POSIX Like | Poll | Not yet 🥺 |
| POSIX Like | Select | Not yet 🥺 |


`liblpc` using interface `Poller` and `Watcher` as abstraction for any backend.

## License

Released under the [MIT License](https://github.com/gen-iot/liblpc/blob/master/License)
