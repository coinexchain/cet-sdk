package msgqueue

import (
	"encoding/json"
	"fmt"

	"github.com/spf13/viper"

	cmn "github.com/tendermint/tendermint/libs/common"
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
}

func NewRegulateWriteDir(dir string) (*RegulateWriteDir, error) {
	var (
		err error
		rgw = &RegulateWriteDir{initChainHeight: viper.GetInt64("genesis_block_height")}
	)
	if rgw.MsgWriter, err = NewDirMsgWriter(dir, getFilePathAndIndex); err != nil {
		return nil, err
	}
	rgw.MsgWriter.(*dirMsgWriter).SetTimeToNewFile(rgw.timeToNewFile())
	return rgw, nil
}

func (r *RegulateWriteDir) timeToNewFile() func(k, v []byte) bool {
	return func(k, v []byte) bool {
		if string(k) == ("height_info") {
			var info NewHeightInfo
			if err := json.Unmarshal(v, &info); err != nil {
				panic(fmt.Sprintf("json unmarshal height_info failed; err: %s\n", err.Error()))
			}
			return info.Height != r.initChainHeight+1 && ((info.Height-r.initChainHeight-1)%FILEHEIGHT) == 0
		}
		return false
	}
}
