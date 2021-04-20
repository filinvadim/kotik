package main

import (
	"fmt"
	"sync"
	"testing"
	"time"
)

type mockRequestor struct{}

func (m *mockRequestor) get(url string) ([]byte, error) {
	time.Sleep(time.Second)
	return []byte(url + "hash"), nil
}

func test_General_Success(t *testing.T) {
	var (
		parallelLimit = 5
		pairs         = make(chan string, parallelLimit)
		urls          = []string{"http://google.com"}
		builder       = NewHashBuilder(&mockRequestor{}, parallelLimit, &sync.WaitGroup{})
	)

	go builder.build(urls, pairs)

	select {
	case p, ok := <-pairs:
		if !ok {
			return
		}
		fmt.Print(p)
	case <-time.After(2 * time.Second):
		t.Fatal("timeout")
	}
}

type mockWaiter struct {
	t *testing.T

	limit int

	doneChan chan struct{}
}

func (m mockWaiter) Add(i int) {
	m.doneChan <- struct{}{}
}

func (m mockWaiter) Done() {
	<-m.doneChan
}

func (m mockWaiter) Wait() {
	time.Sleep(time.Millisecond * 10)
	for {
		if len(m.doneChan) == 0 {
			fmt.Println("WAIT!")
			return
		}
		time.Sleep(time.Second*2)
	}
}

func Test_UrlsMoreThanLimit_Success(t *testing.T) {
	var (
		parallelLimit = 5
		urls          = []string{
			"http://google.com",
			"http://ya.ru",
			"http://yahoo.com",
			"http://google.com",
			"http://amazon.com",
			"http://google.com",
			"http://vk.com",
			"http://google.com",
			"http://instagram.com",
			"http://google.com",
			"http://google.com",
			"http://google.com",
			"http://ya.ru",
			"http://yahoo.com",
			"http://google.com",
			"http://amazon.com",
			"http://google.com",
			"http://vk.com",
			"http://google.com",
			"http://instagram.com",
			"http://google.com",
			"http://google.com",
			"http://google.com",
			"http://ya.ru",
			"http://yahoo.com",
			"http://google.com",
			"http://amazon.com",
			"http://google.com",
			"http://vk.com",
			"http://google.com",
			"http://instagram.com",
			"http://google.com",
			"http://google.com",
			"http://google.com",
			"http://ya.ru",
			"http://yahoo.com",
			"http://google.com",
			"http://amazon.com",
			"http://google.com",
			"http://vk.com",
			"http://google.com",
			"http://instagram.com",
			"http://google.com",
			"http://google.com",
			"http://google.com",
			"http://ya.ru",
			"http://yahoo.com",
			"http://google.com",
			"http://amazon.com",
			"http://google.com",
			"http://vk.com",
			"http://google.com",
			"http://instagram.com",
			"http://google.com",
			"http://google.com",
		}
		pairs  = make(chan string, len(urls))
		mockWg = mockWaiter{
			t:        t,
			limit:    parallelLimit,
			doneChan: make(chan struct{}, parallelLimit),
		}
		builder = NewHashBuilder(&mockRequestor{}, parallelLimit, &mockWg)
	)

	go builder.build(urls, pairs)
	for {
		select {
		case <-time.After(time.Minute):
			t.Fatal("timeout")
		default:
			fmt.Println(len(pairs), "RECEIVED")
			if len(pairs) == len(urls) {
				return
			}
		}
		time.Sleep(time.Second)
	}
}
