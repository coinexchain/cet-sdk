package msgqueue

import (
	"encoding/json"

	"github.com/spf13/viper"

	cmn "github.com/tendermint/tendermint/libs/common"

	ts "github.com/coinexchain/cet-sdk/trade_server"
	"github.com/coinexchain/trade-server/server"
)

const FILEHEIGHT = 10000

type RegulateWriteDir struct {
	*dirMsgWriter
	initChainHeight int64
	currHeight      int64
	pfWork          Worker
	dataWork        Worker
}

func NewRegulateWriteDirAndComponents(dir string) (*RegulateWriteDir, error) {
	rgw := &RegulateWriteDir{initChainHeight: viper.GetInt64("genesis_block_height")}
	w, err := NewDirMsgWriter(dir, getFilePathAndIndex)
	if err != nil {
		return nil, err
	}
	diw := w.(*dirMsgWriter)
	diw.SetTimeToNewFile(rgw.TimeToNewFile())
	rgw.creatPruneAndTradeConsumer(rgw, dir)
	rgw.startWork()
	return rgw, nil
}

func (r *RegulateWriteDir) TimeToNewFile() func(k, v []byte) bool {
	return func(k, v []byte) bool {
		if string(k) == ("height_info") {
			var info NewHeightInfo
			json.Unmarshal(v, &info)
			r.currHeight = info.Height - 1
			return ((info.Height - 1) % FILEHEIGHT) == 0
		}
		return false
	}
}

func (r *RegulateWriteDir) GetHeight() int64 {
	return r.currHeight
}

func getFilePathAndIndex(dir string, height int) (filePath string, fileIndex int, err error) {
	return
}

type NewHeightInfo struct {
	ChainID       string       `json:"chain_id"`
	Height        int64        `json:"height"`
	TimeStamp     int64        `json:"timestamp"`
	LastBlockHash cmn.HexBytes `json:"last_block_hash"`
}

type Worker interface {
	Work()
}

func (r *RegulateWriteDir) creatPruneAndTradeConsumer(ght ExpectGetHeight, dir string) {
	doneHeightCh := make(chan int64)
	pf := NewPruneFile(ght, doneHeightCh, dir)

	conf, err := initConf()
	_ = err
	hub, err := server.CreateHub(conf)
	cdt, err := server.NewConsumerWithDirTail(conf, hub)
	r.dataWork = ts.NewConsumer(cdt, doneHeightCh)
	r.pfWork = pf
}

func (r *RegulateWriteDir) startWork() {
	r.pfWork.Work()
	r.dataWork.Work()
}
