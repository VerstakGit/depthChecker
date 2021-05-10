package depthChecker

import (
	"github.com/verstakGit/go-binance/v2"
	"github.com/verstakGit/go-binance/v2/futures"
	"sync"
	"time"
)

type DepthChecker struct {
	cfg           *Config
	futuresClient *futures.Client
	spotClient    *binance.Client
	wg            sync.WaitGroup
	quitChan      chan struct{}
	alertLevels   map[level]time.Time
}

type level struct {
	symbol     string
	marketType string
	price      float64
}

func NewDepthChecker(cfg *Config, futuresClient *futures.Client, spotClient *binance.Client) *DepthChecker {
	dc := &DepthChecker{
		cfg:           cfg,
		futuresClient: futuresClient,
		spotClient:    spotClient,
		quitChan:      make(chan struct{}),
		alertLevels:   make(map[level]time.Time),
	}

	return dc
}

func (dc *DepthChecker) Run() error {
	for idx := range dc.cfg.Tickers {
		err := dc.startDepthWorker(
			dc.cfg.Tickers[idx].Symbol,
			dc.cfg.Tickers[idx].MarketType,
			dc.cfg.Tickers[idx].LargeOrder,
		)
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
