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
	dir := "testp"
	pf := NewFileDeleter(doneHeightCh, dir)
	pf.Run()

	rwc, err := NewRegulateWriteDir(dir)
	require.Nil(t, err)
	defer os.RemoveAll(dir)

	viper.Set("genesis_block_height", 0)
	key := []byte(`height_info`)
	model := `{"height":%d,"timestamp":36372364939,"last_block_hash":"7FB1B1AB4EEB748652423D72001841073EC00E811E964B1E6FAE9A2E2EC10E07"}`
	require.Nil(t, rwc.WriteKV(key, []byte(fmt.Sprintf(model, 1))))
	require.Nil(t, rwc.WriteKV(key, []byte(fmt.Sprintf(model, 2))))
	require.Nil(t, rwc.WriteKV(key, []byte(fmt.Sprintf(model, 3))))
	require.Nil(t, rwc.WriteKV(key, []byte(fmt.Sprintf(model, 10000))))
	require.Nil(t, rwc.WriteKV(key, []byte(fmt.Sprintf(model, 10001))))
	require.Nil(t, rwc.WriteKV(key, []byte(fmt.Sprintf(model, 10002))))
	require.Nil(t, rwc.WriteKV(key, []byte(fmt.Sprintf(model, 20001))))
	require.Nil(t, rwc.WriteKV(key, []byte(fmt.Sprintf(model, 20002))))

	doneHeightCh <- 20000
	time.Sleep(time.Millisecond)
	files, _ := getAllFilesFromDir(dir)
	require.EqualValues(t, 3, len(files))

	doneHeightCh <- 20001
	time.Sleep(time.Millisecond)
	files, _ = getAllFilesFromDir(dir)
	require.EqualValues(t, 2, len(files))
	require.EqualValues(t, "backup-1", files[0])
	require.EqualValues(t, "backup-2", files[1])

	doneHeightCh <- 30002
	time.Sleep(time.Millisecond)
	files, _ = getAllFilesFromDir(dir)
	fmt.Println(files)
	require.EqualValues(t, 1, len(files))
	require.EqualValues(t, "backup-2", files[0])
}
