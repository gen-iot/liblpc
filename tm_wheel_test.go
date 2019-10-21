package liblpc

import (
	"context"
	"fmt"
	"testing"
	"time"
)

func TestNewTimeWheel(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)
	defer cancel()
	tw := NewTimeWheel(10, 4)
	tw.Execute(ctx)
}

type timeEntry struct {
	t          *testing.T
	createTime time.Time
	refCounter int32
}

func (this *timeEntry) Close() error {
	fmt.Printf(
		"time entry closed, createAt %s, exitAt %s\n",
		this.createTime.Format("2006-01-02 15:04:05"), time.Now().Format("2006-01-02 15:04:05"))
	return nil
}

func (this *timeEntry) GetRefCounter() *int32 {
	return &this.refCounter
}

func newTimeEntry(t *testing.T) BucketEntry {
	return &timeEntry{
		t:          t,
		createTime: time.Now(),
		refCounter: 0,
	}
}

func TestTimeWheel_Entries(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*30)
	defer cancel()
	tw := NewTimeWheel(5, 2)
	go func() {
		tw.Entries() <- newTimeEntry(t)
		time.Sleep(time.Second * 6)
		tw.Entries() <- newTimeEntry(t)
		time.Sleep(time.Second)
		tw.Entries() <- newTimeEntry(t)
	}()
	tw.Execute(ctx)
}
