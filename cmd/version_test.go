package cmd

import (
	"bytes"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestVersionCommand(t *testing.T) {
	originalVersion, originalCommit, originalDate := Version, Commit, BuildDate
	t.Cleanup(func() {
		Version, Commit, BuildDate = originalVersion, originalCommit, originalDate
	})

	Version, Commit, BuildDate = "v1.2.3", "abc123", "2026-08-30T00:00:00Z"
	var output bytes.Buffer
	command := newVersionCmd()
	command.SetOut(&output)

	require.NoError(t, command.Execute())
	assert.Equal(t, "wanderlog v1.2.3\ncommit: abc123\nbuilt: 2026-08-30T00:00:00Z\n", output.String())
}

func TestVersionCommandRejectsArguments(t *testing.T) {
	command := newVersionCmd()
	command.SetArgs([]string{"unexpected"})
	assert.Error(t, command.Execute())
}
