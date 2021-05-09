package main

import (
	"depthChecker/internal/depthChecker"
	"flag"
	"github.com/adshao/go-binance/v2"
	"log"
	"os"
	"os/signal"
	"syscall"
)

func main() {
	cfgPath := flag.String("config", "", "path to config .yaml (check configs example)")
	flag.Parse()
	if *cfgPath == "" {
		log.Println("keys file is not specified")
		return
	}
	cfg, err := depthChecker.ParseConfig(*cfgPath)
	if err != nil {
		log.Println("parse config error", err)
		return
	}

	client := binance.NewFuturesClient(cfg.Keys.APIKey, cfg.Keys.SecretKey)
	dc := depthChecker.NewDepthChecker(cfg, client)
	go func() {
		sig := make(chan os.Signal, 1)
		signal.Notify(sig, syscall.SIGINT, syscall.SIGTERM)
		<-sig
		dc.Shutdown()
	}()

	err = dc.Run()
	if err != nil {
		log.Println("run error", err)
		return
	}
}
