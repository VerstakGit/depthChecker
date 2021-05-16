package main

import (
	"depthChecker/internal/depthChecker"
	"flag"
	"fmt"
	"github.com/verstakGit/go-binance/v2"
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

	futuresClient := binance.NewFuturesClient(cfg.Keys.APIKey, cfg.Keys.SecretKey)
	spotClient := binance.NewClient(cfg.Keys.APIKey, cfg.Keys.SecretKey)
	dc := depthChecker.NewDepthChecker(cfg, futuresClient, spotClient)
	go func() {
		sig := make(chan os.Signal, 1)
		signal.Notify(sig, syscall.SIGINT, syscall.SIGTERM)
		<-sig
		dc.Shutdown()
		fmt.Println("shutdown the process...")
	}()

	err = dc.Run()
	if err != nil {
		log.Println("run error", err)
		return
	}
}
