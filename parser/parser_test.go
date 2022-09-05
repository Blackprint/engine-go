package parser

import (
	"testing"

	"github.com/blackprint/engine-go/test"
	"github.com/davecgh/go-spew/spew"
	"github.com/stretchr/testify/require"
)

func TestParser(t *testing.T) {
	var dst Root

	err := ParseString(string(test.Data), &dst)
	require.NoError(t, err)
	spew.Dump(dst)
}
