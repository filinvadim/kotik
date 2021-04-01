package main

import (
	"bytes"
	"context"
	"crypto/md5"
	"flag"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"os/signal"
	"strings"
	"sync"
	"syscall"
)

const parallelFlagName = "p"

func main() {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	interrupt := make(chan os.Signal, 1)
	signal.Notify(interrupt, syscall.SIGTERM, syscall.SIGINT)

	var parallelLimit int

	flag.IntVar(
		&parallelLimit,
		parallelFlagName,
		10,
		"limit the number of parallel requests. Must be first parameter",
	)

	flag.Parse()

	urls := validateUrls(flag.Args())

	var (
		pairs   = make(chan string, len(urls))
		cli     = NewRequestor(ctx)
		builder = NewHashBuilder(cli, parallelLimit, new(sync.WaitGroup))
		waiter  = make(chan struct{})
	)
	// reader first
	builder.print(waiter, pairs)
	builder.build(urls, pairs)

	select {
	case <-waiter:
	case <-interrupt:
		fmt.Println("interrupted")
	}
}

func validateUrls(tail []string) []string {
	var parsedUrls = make([]string, 0)
	for _, u := range tail {
		if !strings.HasPrefix(u, "http") {
			fmt.Println("unparsed parameter:", u)
			continue
		}
		_, err := url.Parse(u)
		if err != nil {
			fmt.Printf("malformed URL: %s %v", u, err)
			continue
		}
		parsedUrls = append(parsedUrls, u)
	}
	return parsedUrls
}

type HashBuilder interface {
	build(urls []string, pairs chan string)
	print(waiter chan struct{}, pairs chan string)
}

type Waiter interface {
	Add(int)
	Done()
	Wait()
}

type builder struct {
	cli   Requestor
	wg    Waiter
	limit int
}

func NewHashBuilder(cli Requestor, limit int, waiter Waiter) HashBuilder {
	return &builder{
		cli:   cli,
		wg:    waiter,
		limit: limit,
	}
}

func (b *builder) print(waiter chan struct{}, pairs chan string) {
	go func() {
		for p := range pairs {
			fmt.Print(p)
		}
		close(waiter)
	}()
}

func (b *builder) build(urls []string, pairs chan string) {
	defer close(pairs)

	for i, u := range urls {
		if u == "" {
			continue
		}

		if i != 0 && i % b.limit == 0 {
			b.wg.Wait()
		}


		b.wg.Add(1)
		go func(url string) {
			defer b.wg.Done()

			resp, err := b.cli.get(url)
			if err != nil {
				fmt.Printf("request error: %s %v\n", u, err)
				return
			}

			pairs <- fmt.Sprintf("%s %x\n", url, md5.Sum(resp))
		}(u)
	}
	b.wg.Wait()
}

type Requestor interface {
	get(url string) ([]byte, error)
}

type cli struct {
	ctx context.Context
	*http.Client
}

func NewRequestor(ctx context.Context) Requestor {
	return &cli{
		ctx:    ctx,
		Client: http.DefaultClient,
	}
}

func (c *cli) get(url string) ([]byte, error) {
	if c.ctx.Err() != nil {
		return nil, c.ctx.Err()
	}
	req, err := http.NewRequestWithContext(c.ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}

	resp, err := c.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	buf := bytes.NewBuffer(nil)
	_, err = io.Copy(buf, resp.Body)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("code: %d, message: %s \n", resp.StatusCode, buf.String())
	}
	return buf.Bytes(), nil
}
