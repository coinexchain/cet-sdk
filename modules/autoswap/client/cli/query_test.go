package cli

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestQueryParams(t *testing.T) {
	cmd := GetQueryCmd(nil)
	args := []string{
		"params",
	}
	cmd.SetArgs(args)
	err := cmd.Execute()
	assert.Equal(t, nil, err)
	assert.Equal(t, "custom/market/parameters", resultPath)
	assert.Equal(t, nil, resultParam)
}
