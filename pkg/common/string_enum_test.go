package common

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestStringEnumAcceptsAllowedValues(t *testing.T) {
	enum := NewStringEnum("output", "raw", "raw", "json")

	require.Equal(t, "raw", enum.String())
	require.NoError(t, enum.Set("json"))
	require.Equal(t, "json", enum.String())
}

func TestStringEnumRejectsUnsupportedValues(t *testing.T) {
	enum := NewStringEnum("output", "raw", "raw", "json")

	err := enum.Set("yaml")

	require.EqualError(t, err, `unsupported output "yaml" (must be one of: raw, json)`)
	require.Equal(t, "raw", enum.String())
}

func TestStringEnumFlagUsesPrimaryNameInErrors(t *testing.T) {
	flag := StringEnumFlag("output, o", "output format", "raw", "raw", "json")

	err := flag.Value.Set("yaml")

	require.EqualError(t, err, `unsupported output "yaml" (must be one of: raw, json)`)
}
