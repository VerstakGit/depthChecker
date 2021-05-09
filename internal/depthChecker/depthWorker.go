package depthChecker

import (
	"context"
	"github.com/adshao/go-binance/v2/futures"
	"github.com/gen2brain/beeep"
	"log"
	"time"
)

type workerInfo struct {
	events       chan *futures.WsDepthEvent
	symbol       string
	largeOrder   float64
	initialResID int64
	stop         chan struct{}
	done         chan struct{}
	orderBook    map[float64]float64
}

func (dc *DepthChecker) startDepthWorker(symbol string, largeOrder float64) error {
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

	res, err := dc.client.NewDepthService().Limit(1000).Symbol(symbol).Do(context.Background())
	if err != nil {
		return err
	}

	orderBook := make(map[float64]float64, 2000)
	dc.fillOrdersAndAlert(orderBook, res.Bids, largeOrder, symbol)
	dc.fillOrdersAndAlert(orderBook, res.Asks, largeOrder, symbol)
	wInfo := &workerInfo{
		events:       events,
		symbol:       symbol,
		largeOrder:   largeOrder,
		initialResID: res.LastUpdateID,
		stop:         stop,
		done:         done,
		orderBook:    orderBook,
	}

	dc.wg.Add(1)
	go dc.start(wInfo)

	return nil
}

func (dc *DepthChecker) start(wCfg *workerInfo) {
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

func (dc *DepthChecker) fillOrdersAndAlert(orderBook map[float64]float64, orders []futures.PriceLevel, largeOrder float64, symbol string) {
	for idx := range orders {
		p, q, err := orders[idx].Parse()
		if err != nil {
			log.Println("parse depth err", err)
			continue
		}
		orderBook[p] = q
		if p*q >= largeOrder {
			log.Printf("%s:: price = %v, quantity = %v, total sum (USD) = %v", symbol, p, q, int(p*q))
			if dc.cfg.PlaySoundAlert {
				err := beeep.Beep(246, 500)
				if err != nil {
					log.Println("play beep sound error", err)
				}
			}
		}
	}
}
