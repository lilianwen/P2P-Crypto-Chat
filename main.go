package main

import (
	"P2P-Crypto-Chat/common"
	"P2P-Crypto-Chat/peer"
	"flag"
	"os"
	"path/filepath"
)

var defaultLogFilename = "chat.log"

func main() {
	//获取配置文件路径
	cfgFile := flag.String("c", "", "configure file")
	flag.Parse()

	//判断配置文件是否存在
	if _, err := os.Stat(*cfgFile); os.IsNotExist(err) {
		panic(err)
	}

	cfg := common.LoadConfig(*cfgFile)

	//设置日志
	initLogRotator(filepath.Join(cfg.LogDir, defaultLogFilename))
	defer func() {
		if logRotator != nil {
			logRotator.Close()
		}
	}()
	setLogLevels(cfg.LogLevel)

	interrupt := interruptListener()
	if interruptRequested(interrupt) {
		return
	}
	startService(cfg)
	defer func() {
		log.Infof("Gracefully shutting down the server...")
		stopService()
		log.Infof("Server shutdown complete")
	}()

	<- interrupt
}

func startService(cfg common.Config) {
	peer.Start(cfg)
}

func stopService(){
	peer.Stop()
}
