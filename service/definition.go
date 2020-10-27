package service

import "context"

type UrlDownloader interface {
	Download(ctx context.Context, urls []string) ([]DownloaderResponse, error)
}
