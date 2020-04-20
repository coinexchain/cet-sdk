package msgqueue

import (
	"os"

	"github.com/prometheus/common/log"

	"github.com/coinexchain/trade-server/core"
	"github.com/coinexchain/trade-server/server"
	toml "github.com/pelletier/go-toml"
	"github.com/spf13/viper"
	cfg "github.com/tendermint/tendermint/config"
)

type MemWriteConsumer struct {
	*server.TradeConsumerWithMemBuf
}

func NewMemWriteConsumer() (*MemWriteConsumer, error) {
	var (
		err  error
		hub  *core.Hub
		conf *toml.Tree
		tc   *server.TradeConsumerWithMemBuf
	)

	if conf, err = initConf(); err != nil {
		log.Errorf("Load file failed: %s\n", err.Error())
		return nil, err
	}
	if hub, err = server.CreateHub(conf); err != nil {
		log.Errorf("Init Hub failed: %s\n", err.Error())
		return nil, err
	}
	if tc, err = server.NewConsumerWithMemBuf(conf, hub); err != nil {
		log.Error("Init TradeConsumerWithMemBuf failed: %s\n", err.Error())
		return nil, err
	}

	return &MemWriteConsumer{TradeConsumerWithMemBuf: tc}, nil
}

func initConf() (*toml.Tree, error) {
	conf := cfg.DefaultConfig()
	err := viper.Unmarshal(conf)
	if err != nil {
		return nil, err
	}
	filePath := conf.RootDir + "config/trade-server.toml"
	info, err := os.Stat(filePath)
	if err != nil || info.IsDir() {
		return nil, err
	}
	return toml.LoadFile(filePath)
}

func (m *MemWriteConsumer) WriteKV(k, v []byte) error {
	m.PutMsg(k, v)
	return nil
}

func (m *MemWriteConsumer) Close() error {
	m.TradeConsumerWithMemBuf.Close()
	return nil
}
