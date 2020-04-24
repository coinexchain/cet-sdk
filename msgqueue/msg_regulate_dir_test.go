package msgqueue

import (
	"fmt"
	"os"
	"testing"

	"github.com/spf13/viper"

	"github.com/stretchr/testify/require"
)

func TestNewRegulateWriteDirAndComponents(t *testing.T) {
	rwc, err := NewRegulateWriteDir("test")
	defer os.RemoveAll("test")
	require.Nil(t, err)

	viper.Set("genesis_block_height", 1)
	key := []byte(`height_info`)
	model := `{"height":%d,"timestamp":36372364939,"last_block_hash":"7FB1B1AB4EEB748652423D72001841073EC00E811E964B1E6FAE9A2E2EC10E07"}`
	rwc.WriteKV(key, []byte(fmt.Sprintf(model, 1)))
	rwc.WriteKV(key, []byte(fmt.Sprintf(model, 2)))
	rwc.WriteKV(key, []byte(fmt.Sprintf(model, 3)))
	rwc.WriteKV(key, []byte(fmt.Sprintf(model, 10000)))
	files, err := getAllFilesFromDir("test")
	require.Nil(t, err)
	require.EqualValues(t, 1, len(files))
	require.EqualValues(t, "backup-0", files[0])

	rwc.WriteKV(key, []byte(fmt.Sprintf(model, 10001)))
	rwc.WriteKV(key, []byte(fmt.Sprintf(model, 10002)))
	rwc.WriteKV(key, []byte(fmt.Sprintf(model, 20000)))
	files, err = getAllFilesFromDir("test")
	require.Nil(t, err)
	require.EqualValues(t, 2, len(files))
	require.EqualValues(t, "backup-0", files[0])
	require.EqualValues(t, "backup-1", files[1])

	rwc.WriteKV(key, []byte(fmt.Sprintf(model, 20001)))
	rwc.WriteKV(key, []byte(fmt.Sprintf(model, 20002)))
	rwc.WriteKV(key, []byte(fmt.Sprintf(model, 30000)))
	files, err = getAllFilesFromDir("test")
	require.Nil(t, err)
	require.EqualValues(t, 3, len(files))
	require.EqualValues(t, "backup-0", files[0])
	require.EqualValues(t, "backup-1", files[1])
	require.EqualValues(t, "backup-2", files[2])

}
