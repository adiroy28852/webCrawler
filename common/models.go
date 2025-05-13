package common

import (
	"sync"
	"time"
)

type CLIFlags struct {
	seedUrls           []string
	numWorkers         int32
	crawlDelay         time.Duration
	DBConnectionString string
	userAgent          string
}

type ConfigManager struct {
	seedUrls           []string
	numWorkers         int32
	crawlDelay         time.Duration
	DBConnectionString string
	userAgent          string
}

type FetchedPageData struct {
	URL   string
	body  []byte
	error error
}

type PageStorageData struct {
	URL   string
	title string
	error error
}

type UrlManager struct {
	mu              sync.Mutex
	queue           []string
	visited         map[string]bool
	urlChannel      chan string
	activeWorkers   sync.WaitGroup
	shutDownChannel chan struct{}
	done            bool
}
