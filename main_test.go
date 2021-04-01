package main

import (
	"fmt"
	"sync"
	"testing"
	"time"
)

type mockRequestor struct {}

func (m *mockRequestor) get(url string) ([]byte, error) {
	time.Sleep(time.Millisecond * 10)
	return []byte{}, nil
}

func Test_General_Success(t *testing.T) {
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

func (m mockWaiter) Add(i int) {}

func (m mockWaiter) Done() {
	m.doneChan <- struct{}{}
}

// Wait must be called when parallel limit has reached or fail
func (m mockWaiter) Wait() {
	counter := 0
	for range m.doneChan {
		if counter > (m.limit - 1) {
			m.t.Fatal("limiting the number of parallel requests is failed")
		}
		counter++
	}
}

func Test_UrlsMoreThanLimit_Success(t *testing.T) {
	var (
		parallelLimit = 5
		pairs         = make(chan string, parallelLimit)
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
		}
		mockWg = mockWaiter{
			t:        t,
			limit:    parallelLimit,
			doneChan: make(chan struct{}, parallelLimit),
		}
		builder        = NewHashBuilder(&mockRequestor{}, parallelLimit, &mockWg)
		requestCounter int
	)
	defer close(mockWg.doneChan)

	go builder.build(urls, pairs)
	for {
		select {
		case <-pairs:
			// only first limited batch of pairs that were received before Wait call
			requestCounter++

			if requestCounter == parallelLimit {
				return
			}

		case <-time.After(5 * time.Second):
			t.Fatal("timeout")
		}
	}
}
