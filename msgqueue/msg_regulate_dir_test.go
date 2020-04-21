package msgqueue

import (
	"os"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestNewRegulateWriteDirAndComponents(t *testing.T) {
	rwc, err := NewRegulateWriteDir("test")
	defer os.RemoveAll("test")
	require.Nil(t, err)
	_ = rwc

}
