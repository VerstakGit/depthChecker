package depthChecker

import (
	"context"
	"github.com/verstakGit/go-binance/v2/futures"
	"log"
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

	res, err := dc.futuresClient.NewDepthService().Limit(1000).Symbol(symbol).Do(context.Background())
	if err != nil {
		return err
	}

	orderBook := make(map[float64]float64, 2000)
	dc.fillOrdersAndAlert(orderBook, res.Bids, largeOrder, symbol)
	dc.fillOrdersAndAlert(orderBook, res.Asks, largeOrder, symbol)
	wInfo := &futuresInfo{
		events:       events,
		symbol:       symbol,
		largeOrder:   largeOrder,
		initialResID: res.LastUpdateID,
		stop:         stop,
		done:         done,
		orderBook:    orderBook,
	}

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
			dc.fillOrdersAndAlert(wCfg.orderBook, event.Bids, wCfg.largeOrder, wCfg.symbol)
			dc.fillOrdersAndAlert(wCfg.orderBook, event.Asks, wCfg.largeOrder, wCfg.symbol)
			prevID = event.LastUpdateID
		case <-dc.quitChan:
			return
		}
	}

}
