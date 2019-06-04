package backend

import (
	"fmt"
	"os"
	"testing"
	"time"
)

func panicIfError(err error) {
	if err != nil {
		panic(err)
	}
}

func pollerEvtTriger(poller *Epoll) {
	time.Sleep(time.Second)
	for {
		poller.Notify(func() {
			fmt.Println("im pollerEvtTriger")
		})
		time.Sleep(time.Second)
	}
}

func pollerTriger(poller *Epoll) {
	time.Sleep(time.Second)
	tmpFile, err := os.OpenFile("tmp", os.O_CREATE|os.O_RDWR, 0755)
	panicIfError(err)
	err = poller.AddFd(int(tmpFile.Fd()), 0, nil)
	panicIfError(err)
	for {
		_, _ = tmpFile.WriteString(fmt.Sprintf("%s\n", time.Now().String()))
		time.Sleep(time.Second)
	}
}

func TestPoll(t *testing.T) {
	poller, err := NewPoll()
	panicIfError(err)
	go pollerEvtTriger(poller)
	err = poller.Wait()
	panicIfError(err)
}
