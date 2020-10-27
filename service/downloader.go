package service

import (
	"context"
	"encoding/json"
	"errors"
	"io/ioutil"
	"net/http"
	"time"
)

func NewUrlDownloader(out uint, timeout time.Duration) UrlDownloader {
	d := &Downloader{
		requestTimeout: timeout,
		maxWorkers:     out,
	}

	return d
}

func (d Downloader) Download(ctx context.Context, urls []string) ([]DownloaderResponse, error) {
	newCtx, cancel := context.WithCancel(ctx)
	resp := make(chan DownloaderResponse)
	defer func() {
		close(resp)
	}()

	expectedURLs := len(urls)

	go d.queueURLs(newCtx, urls, resp)

	data, err := d.responseWaiter(newCtx, cancel, expectedURLs, resp)
	if err != nil {
		return nil, err
	}

	return data, nil
}

type downloaderRequest struct {
	ctx context.Context
	out chan<- DownloaderResponse
	url string
}

type Downloader struct {
	requestTimeout time.Duration
	maxWorkers     uint
}

type DownloaderResponse struct {
	// Response - json
	URL      string
	Response interface{}
	Err      error
}

func (d Downloader) makeRequest(request downloaderRequest) ([]byte, error) {
	cl := http.Client{Timeout: d.requestTimeout}

	req, err := http.NewRequestWithContext(request.ctx, "GET", request.url, nil)
	if err != nil {
		return nil, err
	}

	errC := make(chan error, 1)
	respC := make(chan *http.Response, 1)
	go func() {
		resp, err := cl.Do(req)
		if err != nil {
			errC <- err
		}

		respC <- resp
	}()

	select {
	case <-request.ctx.Done():
	case resp := <-respC:
		data, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			return nil, err
		}

		return data, nil
	case err := <-errC:
		return nil, err
	}

	return nil, nil
}

func (d Downloader) worker(dr <-chan downloaderRequest) {
	for {
		select {
		case v, ok := <-dr:
			if !ok {
				return
			}

			resp, err := d.makeRequest(v)

			go func() {
				if err != nil {
					v.out <- DownloaderResponse{
						Err: err,
						URL: v.url,
					}

					return
				}

				var dd interface{}
				err = json.Unmarshal(resp, &dd)
				if err != nil {
					v.out <- DownloaderResponse{
						Err: err,
						URL: v.url,
					}

					return
				}

				v.out <- DownloaderResponse{
					Response: dd,
					URL:      v.url,
				}
			}()
		}
	}
}

func (d Downloader) closeChannels(dt []chan downloaderRequest) {
	for _, v := range dt {
		close(v)
	}
}

func (d Downloader) queueURLs(ctx context.Context, urls []string, out chan<- DownloaderResponse) {
	var (
		head       = len(urls) - 1
		currentUrl = downloaderRequest{
			out: out,
			url: urls[head],
			ctx: ctx,
		}
	)

	workers := make([]chan downloaderRequest, d.maxWorkers)

	var i uint
	for i = 0; i < d.maxWorkers; i++ {
		workers[i] = make(chan downloaderRequest)

		go d.worker(workers[i])
	}

	for {
		for _, v := range workers {
			select {
			case <-ctx.Done():
				d.closeChannels(workers)

				return
			default:
			}

			select {
			case v <- currentUrl:
				head--

				if head == -1 {
					d.closeChannels(workers)

					return
				}

				currentUrl = downloaderRequest{
					out: out,
					url: urls[head],
					ctx: ctx,
				}
			default:
			}
		}
	}
}

func (d *Downloader) responseWaiter(ctx context.Context, cancel context.CancelFunc, expected int, resp chan DownloaderResponse) ([]DownloaderResponse, error) {
	responses := make([]DownloaderResponse, 0, expected)
	cURLs := 0

	for {
		select {
		case <-ctx.Done():
			return nil, errors.New("request cancelled")
		case data := <-resp:
			responses = append(responses, data)
			cURLs++

			if data.Err != nil {
				cancel()

				return nil, data.Err
			}

			if cURLs == expected {
				return responses, nil
			}
		}
	}
}
