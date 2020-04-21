package trade_server

import (
	"time"

	"github.com/coinexchain/trade-server/server"
)

type Consumer struct {
	*server.TradeConsumerWithDirTail
	doneHeightCh chan<- int64
	dirName      string
}

func NewConsumer(tradeConsumerWithDirTail *server.TradeConsumerWithDirTail, doneHeightCh chan<- int64) *Consumer {
	return &Consumer{TradeConsumerWithDirTail: tradeConsumerWithDirTail, doneHeightCh: doneHeightCh}
}

func (c *Consumer) work() {
	c.Consume()
	go c.tickDoneHeight()
}

func (c *Consumer) tickDoneHeight() {
	tick := time.NewTicker(60 * time.Second)
	defer tick.Stop()

	for {
		<-tick.C
		doneHeight := c.GetLeastHeight()
		if doneHeight != 0 {
			c.doneHeightCh <- doneHeight
		}
	}
}
