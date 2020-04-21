package trade_server

import (
	"time"

	"github.com/coinexchain/trade-server/server"
)

type Consumer struct {
	server.Consumer
	doneHeightCh chan<- int64
	dirName      string
}

func NewConsumer(tradeConsumerWithDirTail server.Consumer, doneHeightCh chan<- int64) *Consumer {
	return &Consumer{Consumer: tradeConsumerWithDirTail, doneHeightCh: doneHeightCh}
}

func (c *Consumer) Work() {
	c.Consume()
	go c.tickDoneHeight()
}

func (c *Consumer) tickDoneHeight() {
	tick := time.NewTicker(60 * time.Second)
	defer tick.Stop()

	for {
		<-tick.C
		tcd := c.Consumer.(*server.TradeConsumerWithDirTail)
		doneHeight := tcd.GetDumpHeight()
		if doneHeight != 0 {
			c.doneHeightCh <- doneHeight
		}
	}
}
