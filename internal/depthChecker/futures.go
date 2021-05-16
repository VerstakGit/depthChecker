package depthChecker

import (
	"context"
	"depthChecker/internal"
	"fmt"
	"github.com/verstakGit/go-binance/v2/futures"
	"log"
	"sort"
	"strconv"
	"time"
)

type futuresInfo struct {
	events       chan *futures.WsDepthEvent
	symbol       string
	largeOrder   float64
	initialResID int64
	stop         chan struct{}
	done         chan struct{}
	orderBook    map[float64]float64
}

func (dc *DepthChecker) startFuturesDepth(symbol string, largeOrder float64) error {
	events := make(chan *futures.WsDepthEvent, 100)
	depthHandler := func(event *futures.WsDepthEvent) {
		events <- event
	}
	errHandler := func(err error) {
		log.Println("depth websocket error", err)
	}

	done, stop, err := futures.WsDiffDepthServeWithRate(symbol, 500*time.Millisecond, depthHandler, errHandler)
	if err != nil {
		return err
	}

	res, err := dc.futuresClient.NewDepthService().Limit(dc.cfg.FuturesDepthLimit).Symbol(symbol).Do(context.Background())
	if err != nil {
		return err
	}

	orderBook := make(map[float64]float64, 2000)
	dc.fillOrdersAndAlert(orderBook, res.Bids, largeOrder, symbol, "futures")
	dc.fillOrdersAndAlert(orderBook, res.Asks, largeOrder, symbol, "futures")
	wInfo := &futuresInfo{
		events:       events,
		symbol:       symbol,
		largeOrder:   largeOrder,
		initialResID: res.LastUpdateID,
		stop:         stop,
		done:         done,
		orderBook:    orderBook,
	}

	fmt.Printf("starting futures %s worker\n", symbol)
	dc.wg.Add(1)
	go dc.startFuturesWorker(wInfo)

	return nil
}

func (dc *DepthChecker) startFuturesWorker(wCfg *futuresInfo) {
	defer func() {
		wCfg.stop <- struct{}{}
		<-wCfg.done
		dc.wg.Done()
	}()

	prevID := int64(0)
	for {
		select {
		case event := <-wCfg.events:
			if event.LastUpdateID <= wCfg.initialResID {
				continue
			}
			if prevID != 0 && prevID != event.PrevLastUpdateID {
				log.Println("order book desync", wCfg.symbol)
			}
			dc.fillOrdersAndAlert(wCfg.orderBook, event.Bids, wCfg.largeOrder, wCfg.symbol, "futures")
			dc.fillOrdersAndAlert(wCfg.orderBook, event.Asks, wCfg.largeOrder, wCfg.symbol, "futures")
			prevID = event.LastUpdateID
		case <-dc.quitChan:
			return
		}
	}

}

func (dc *DepthChecker) startSymbolsByVol() error {
	stats, err := dc.futuresClient.NewListPriceChangeStatsService().Do(context.Background())
	if err != nil {
		return err
	}

	sort.Slice(stats, func(i, j int) bool {
		volI, _ := strconv.ParseFloat(stats[i].QuoteVolume, 64)
		volJ, _ := strconv.ParseFloat(stats[j].QuoteVolume, 64)
		return volI < volJ
	})
	cnt := 0
	for idx := len(stats) - 1; idx >= 0; idx-- {
		if cnt >= dc.cfg.VolTickers.SymbolCnt {
			return nil
		}
		if internal.Contains(dc.cfg.VolTickers.SymbolBanList, stats[idx].Symbol) {
			continue
		}

		err = dc.startFuturesDepth(stats[idx].Symbol, dc.cfg.VolTickers.LargeOrder)
		if err != nil {
			return err
		}
		cnt++
	}
	return nil
}
