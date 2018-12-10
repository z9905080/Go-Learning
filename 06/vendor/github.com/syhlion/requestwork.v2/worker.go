package requestwork

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"net/url"
	"time"
)

type job struct {
	ctx     context.Context
	req     *http.Request
	handler func(resp *http.Response, err error) error

	end chan error
}

type result struct {
	resp *http.Response
	err  error
}

//DefaultMaxIdleConnPerHost max idle
const DefaultMaxIdleConnPerHost = 20

//New return http worker
func New(threads int) *Worker {

	tr := &http.Transport{
		Proxy:               NoProxyAllowed,
		MaxIdleConnsPerHost: threads * DefaultMaxIdleConnPerHost,
		Dial:                PrintLocalDial,
	}
	client := &http.Client{
		Transport: tr,
		Timeout:   time.Second * 60,
	}
	w := &Worker{
		jobQuene: make(chan *job),
		threads:  threads,
		tr:       tr,
		client:   client,
	}

	go w.start()
	return w

}

func PrintLocalDial(network, addr string) (net.Conn, error) {
	dial := net.Dialer{
		Timeout:   30 * time.Second,
		KeepAlive: 30 * time.Second,
	}

	conn, err := dial.Dial(network, addr)
	if err != nil {
		return conn, err
	}

	fmt.Println("connect done, use", conn.LocalAddr().String())

	return conn, err
}

//NoProxyAllowed no proxy
func NoProxyAllowed(request *http.Request) (*url.URL, error) {
	return nil, nil
}

//Worker instance
type Worker struct {
	jobQuene chan *job
	threads  int
	tr       *http.Transport
	client   *http.Client
}

//Execute exec http request
func (w *Worker) Execute(ctx context.Context, req *http.Request, h func(resp *http.Response, err error) error) (err error) {

	j := &job{ctx, req, h, make(chan error)}
	w.jobQuene <- j
	return <-j.end

}

func (w *Worker) run() {
	for j := range w.jobQuene {
		c := make(chan error, 1)
		go func() {
			c <- j.handler(w.client.Do(j.req))
		}()
		select {
		case <-j.ctx.Done():
			w.tr.CancelRequest(j.req)
			j.end <- j.ctx.Err()
		case err := <-c:
			j.end <- err
		}
	}

}

func (w *Worker) start() {

	for i := 0; i < w.threads; i++ {
		go w.run()
	}

}
