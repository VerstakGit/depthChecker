package depthChecker

import (
	"github.com/adshao/go-binance/v2/futures"
	"sync"
)

type DepthChecker struct {
	cfg      *Config
	client   *futures.Client
	wg       sync.WaitGroup
	quitChan chan struct{}
}

func NewDepthChecker(cfg *Config, client *futures.Client) *DepthChecker {
	dc := &DepthChecker{
		cfg:      cfg,
		client:   client,
		quitChan: make(chan struct{}),
	}

	return dc
}

func (dc *DepthChecker) Run() error {
	for idx := range dc.cfg.Tickers {
		err := dc.startDepthWorker(dc.cfg.Tickers[idx].Symbol, dc.cfg.Tickers[idx].LargeOrder)
		if err != nil {
			return err
		}
	}

	dc.wg.Wait()
	return nil
}

func (dc *DepthChecker) Shutdown() {
	close(dc.quitChan)
}
