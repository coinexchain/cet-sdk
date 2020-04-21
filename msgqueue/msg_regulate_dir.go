package msgqueue

import (
	"encoding/json"

	cmn "github.com/tendermint/tendermint/libs/common"
)

const FILEHEIGHT = 10000

type RegulateWriteDir struct {
	*dirMsgWriter
	initChainHeight int64
}

func NewRegulateWriteDir(dir string, initChainHeight int64) (*RegulateWriteDir, error) {
	rgw := &RegulateWriteDir{initChainHeight: initChainHeight}
	w, err := NewDirMsgWriter(dir, getFilePathAndIndex)
	if err != nil {
		return nil, err
	}
	diw := w.(*dirMsgWriter)
	diw.SetTimeToNewFile(rgw.TimeToNewFile())
	return rgw, nil
}

func (r *RegulateWriteDir) TimeToNewFile() func(k, v []byte) bool {
	return func(k, v []byte) bool {
		if string(k) == ("height_info") {
			var info NewHeightInfo
			json.Unmarshal(v, &info)
			return ((info.Height - 1) % FILEHEIGHT) == 0
		}
		return false
	}
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
