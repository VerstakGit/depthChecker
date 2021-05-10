package depthChecker

import (
	"github.com/gen2brain/beeep"
	"github.com/verstakGit/go-binance/v2/common"
	"log"
	"time"
)

const (
	spotMarket    = "s"
	futuresMarket = "f"
)

func (dc *DepthChecker) startDepthWorker(symbol, marketType string, largeOrder float64) error {
	var err error
	switch marketType {
	case spotMarket:
		err = dc.startSpotDepth(symbol, largeOrder)
	case futuresMarket:
		err = dc.startFuturesDepth(symbol, largeOrder)
	default:
		err = ErrUnknownMarketType{Symbol: symbol}
	}

	return err
}

func (dc *DepthChecker) fillOrdersAndAlert(orderBook map[float64]float64, orders []common.PriceLevel, largeOrder float64, symbol, marketType string) {
	for idx := range orders {
		p, q, err := orders[idx].Parse()
		if err != nil {
			log.Println("parse depth err", err)
			continue
		}
		orderBook[p] = q
		if p*q >= largeOrder {
			level := level{
				symbol:     symbol,
				marketType: marketType,
				price:      p,
			}
			prevTime, ok := dc.alertLevels[level]
			curTime := time.Now()
			if ok && prevTime.Add(dc.cfg.AlertCooldown).After(curTime) {
				continue
			}
			dc.alertLevels[level] = curTime
			log.Printf("%s %s:: price = %v, quantity = %v, total sum (USD) = %v", symbol, marketType, p, q, int(p*q))
			if dc.cfg.PlaySoundAlert {
				err := beeep.Beep(246, 500)
				if err != nil {
					log.Println("play beep sound error", err)
				}
			}
		}
	}
}
