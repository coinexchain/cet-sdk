package msgqueue

import (
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/spf13/viper"
)

func TestNewPruneFile(t *testing.T) {
	doneHeightCh := make(chan int64)
	pf := NewPruneFile(doneHeightCh, "test")
	pf.Work()

	rwc, err := NewRegulateWriteDir("test")
	require.Nil(t, err)
	defer os.RemoveAll("test")

	viper.Set("genesis_block_height", 1)
	key := []byte(`height_info`)
	model := `{"height":%d,"timestamp":36372364939,"last_block_hash":"7FB1B1AB4EEB748652423D72001841073EC00E811E964B1E6FAE9A2E2EC10E07"}`
	rwc.WriteKV(key, []byte(fmt.Sprintf(model, 1)))
	rwc.WriteKV(key, []byte(fmt.Sprintf(model, 2)))
	rwc.WriteKV(key, []byte(fmt.Sprintf(model, 3)))
	rwc.WriteKV(key, []byte(fmt.Sprintf(model, 10001)))
	rwc.WriteKV(key, []byte(fmt.Sprintf(model, 10002)))
	rwc.WriteKV(key, []byte(fmt.Sprintf(model, 20001)))
	rwc.WriteKV(key, []byte(fmt.Sprintf(model, 20002)))

	doneHeightCh <- 20000
	time.Sleep(time.Millisecond)
	files, _ := getAllFilesFromDir("test")
	require.EqualValues(t, 3, len(files))

	doneHeightCh <- 20001
	time.Sleep(time.Millisecond)
	files, _ = getAllFilesFromDir("test")
	require.EqualValues(t, 2, len(files))
	require.EqualValues(t, "backup-1", files[0])
	require.EqualValues(t, "backup-2", files[1])

	doneHeightCh <- 30001
	time.Sleep(time.Millisecond)
	files, _ = getAllFilesFromDir("test")
	require.EqualValues(t, 1, len(files))
	require.EqualValues(t, "backup-2", files[0])
}
