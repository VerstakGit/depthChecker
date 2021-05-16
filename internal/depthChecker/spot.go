package depthChecker

import (
	"context"
	"fmt"
	"github.com/verstakGit/go-binance/v2"
	"log"
)

type spotInfo struct {
	events       chan *binance.WsDepthEvent
	symbol       string
	largeOrder   float64
	initialResID int64
	stop         chan struct{}
	done         chan struct{}
	orderBook    map[float64]float64
}

func (dc *DepthChecker) startSpotDepth(symbol string, largeOrder float64) error {
	events := make(chan *binance.WsDepthEvent, 100)
	depthHandler := func(event *binance.WsDepthEvent) {
		events <- event
	}
	errHandler := func(err error) {
		log.Println("depth websocket error", err)
	}

	done, stop, err := binance.WsDepthServe(symbol, depthHandler, errHandler)
	if err != nil {
		return err
	}

	res, err := dc.spotClient.NewDepthService().Limit(dc.cfg.SpotDepthLimit).Symbol(symbol).Do(context.Background())
	if err != nil {
		return err
	}

	orderBook := make(map[float64]float64, 5000)
	dc.fillOrdersAndAlert(orderBook, res.Bids, largeOrder, symbol, "spot")
	dc.fillOrdersAndAlert(orderBook, res.Asks, largeOrder, symbol, "spot")
	wInfo := &spotInfo{
		events:       events,
		symbol:       symbol,
		largeOrder:   largeOrder,
		initialResID: res.LastUpdateID,
		stop:         stop,
		done:         done,
		orderBook:    orderBook,
	}

	fmt.Printf("starting spot %s worker\n", symbol)
	dc.wg.Add(1)
	go dc.startSpotWorker(wInfo)

	return nil
}

func (dc *DepthChecker) startSpotWorker(wCfg *spotInfo) {
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
			if prevID != 0 && prevID+1 != event.FirstUpdateID {
				log.Println("order book desync", wCfg.symbol)
			}
			dc.fillOrdersAndAlert(wCfg.orderBook, event.Bids, wCfg.largeOrder, wCfg.symbol, "spot")
			dc.fillOrdersAndAlert(wCfg.orderBook, event.Asks, wCfg.largeOrder, wCfg.symbol, "spot")
			prevID = event.LastUpdateID
		case <-dc.quitChan:
			return
		}
	}

}
