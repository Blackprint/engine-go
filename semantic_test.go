package engine

import (
	"testing"

	"github.com/blackprint/engine-go/internal/test"
	"github.com/blackprint/engine-go/parser"
	"github.com/davecgh/go-spew/spew"
	"github.com/stretchr/testify/require"
)

func TestSemantic(t *testing.T) {
	var dst parser.Root
	var err error
	err = parser.ParseString(string(test.DataRaw), &dst)
	require.NoError(t, err)
	sem := newSemantic()
	err = sem.process(dst)
	require.NoError(t, err)

	spew.Dump(sem)
}
