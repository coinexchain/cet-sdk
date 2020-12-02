package autoswap

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestX(t *testing.T) {
	mb := AppModuleBasic{}
	gene := mb.DefaultGenesis()
	err := mb.ValidateGenesis(gene)
	require.NoError(t, err)
}
