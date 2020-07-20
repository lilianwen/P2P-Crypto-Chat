package main

import (
	"P2P-Crypto-Chat/p2p"
	"P2P-Crypto-Chat/peer"
	"fmt"
	"github.com/btcsuite/btclog"
	"github.com/jrick/logrotate/rotator"
	"os"
	"path/filepath"
)

type logWriter struct{}

func (logWriter) Write(p []byte) (n int, err error) {
	_, _ = os.Stdout.Write(p)
	_, _ = logRotator.Write(p)
	return len(p), nil
}

var (
	backendLog = btclog.NewBackend(logWriter{}, btclog.WithFlags(btclog.Lshortfile))//增加文件名和行号
	logRotator *rotator.Rotator
	log = backendLog.Logger("MAIN")
	p2pLogger = backendLog.Logger("P2P")
	peerLogger = backendLog.Logger("PEER")
)

func init() {
	p2p.UseLogger(p2pLogger)
	peer.UseLogger(peerLogger)
}

var subsystemLoggers = map[string]btclog.Logger{
	"PEER": peerLogger,
	"P2P":  p2pLogger,
	"MAIN": log,
}


func initLogRotator(logFile string) {
	logDir, _ := filepath.Split(logFile)
	err := os.MkdirAll(logDir, 0700)
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to create log directory: %v\n", err)
		os.Exit(1)
	}
	r, err := rotator.New(logFile, 10*1024, false, 3)
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to create file rotator: %v\n", err)
		os.Exit(1)
	}

	logRotator = r
}

func setLogLevels(logLevel string) {
	// Configure all sub-systems with the new logging level.  Dynamically
	// create loggers as needed.
	for subsystemID := range subsystemLoggers {
		setLogLevel(subsystemID, logLevel)
	}
}

func setLogLevel(subsystemID string, logLevel string) {
	// Ignore invalid subsystems.
	logger, ok := subsystemLoggers[subsystemID]
	if !ok {
		return
	}

	// Defaults to info if the log level is invalid.
	level, _ := btclog.LevelFromString(logLevel)
	logger.SetLevel(level)
}


