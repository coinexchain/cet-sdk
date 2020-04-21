package msgqueue

import (
	"encoding/json"

	"github.com/spf13/viper"

	cmn "github.com/tendermint/tendermint/libs/common"

	ts "github.com/coinexchain/cet-sdk/trade_server"
	"github.com/coinexchain/trade-server/server"
)

const FILEHEIGHT = 10000

type NewHeightInfo struct {
	ChainID       string       `json:"chain_id"`
	Height        int64        `json:"height"`
	TimeStamp     int64        `json:"timestamp"`
	LastBlockHash cmn.HexBytes `json:"last_block_hash"`
}

type Worker interface {
	Work()
}

type RegulateWriteDir struct {
	MsgWriter
	initChainHeight int64
	pfWork          Worker
	dataWork        Worker
}

func NewRegulateWriteDirAndComponents(dir string) (*RegulateWriteDir, error) {
	var (
		err error
		rgw = &RegulateWriteDir{initChainHeight: viper.GetInt64("genesis_block_height")}
	)
	if rgw.MsgWriter, err = NewDirMsgWriter(dir, getFilePathAndIndex); err != nil {
		return nil, err
	}
	rgw.MsgWriter.(*dirMsgWriter).SetTimeToNewFile(rgw.timeToNewFile())
	if err = rgw.creatPruneAndTradeConsumer(dir); err != nil {
		return nil, err
	}
	rgw.startWork()
	return rgw, nil
}

func (r *RegulateWriteDir) timeToNewFile() func(k, v []byte) bool {
	return func(k, v []byte) bool {
		if string(k) == ("height_info") {
			var info NewHeightInfo
			json.Unmarshal(v, &info)
			return ((info.Height - r.initChainHeight - 1) % FILEHEIGHT) == 0
		}
		return false
	}
}

func (r *RegulateWriteDir) creatPruneAndTradeConsumer(dir string) error {
	doneHeightCh := make(chan int64)
	r.pfWork = NewPruneFile(doneHeightCh, dir)

	conf, err := initConf()
	if err != nil {
		return err
	}
	hub, err := server.CreateHub(conf)
	if err != nil {
		return err
	}
	cdt, err := server.NewConsumerWithDirTail(conf, hub)
	if err != nil {
		return err
	}
	r.dataWork = ts.NewConsumer(cdt, doneHeightCh)
	return nil
}

func (r *RegulateWriteDir) startWork() {
	r.pfWork.Work()
	r.dataWork.Work()
}
