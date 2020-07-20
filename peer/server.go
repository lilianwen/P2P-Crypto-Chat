package peer

import "P2P-Crypto-Chat/common"

var defaultPeer *Peer

func Start(cfg common.Config) {
	log.Info("start service")

	defaultPeer = NewPeer(&cfg)
	defaultPeer.Start()
}

func Stop() {
	log.Info("stop service")
	defaultPeer.Stop()
}