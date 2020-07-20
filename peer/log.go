package peer

import "github.com/btcsuite/btclog"

var log btclog.Logger

func init () {
	log = btclog.Disabled
}

func UseLogger(lger btclog.Logger) {
	log = lger
}