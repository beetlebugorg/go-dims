package core

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/require"
)

// A malformed setting used to be printed and ignored, so the service started
// with silent defaults. Validate reports it instead.
func TestValidateReportsStartupErrors(t *testing.T) {
	saved := startupErrors
	t.Cleanup(func() { startupErrors = saved })

	startupErrors = []error{errors.New(`strconv.ParseInt: parsing "abc"`)}

	config := *ReadConfig()
	err := config.Validate()

	require.Error(t, err)
	require.Contains(t, err.Error(), "invalid configuration")
	require.Contains(t, err.Error(), "abc")
}

func TestValidateAcceptsAWorkingConfiguration(t *testing.T) {
	saved := startupErrors
	t.Cleanup(func() { startupErrors = saved })
	startupErrors = nil

	config := *ReadConfig()
	config.OutputFormat.Default = ""

	require.NoError(t, config.Validate())
}

func TestValidateRejectsAnUnknownDefaultFormat(t *testing.T) {
	saved := startupErrors
	t.Cleanup(func() { startupErrors = saved })
	startupErrors = nil

	config := *ReadConfig()
	config.OutputFormat.Default = "WEBP"

	require.Error(t, config.Validate(), "the map is keyed in lower case")
}

func TestRecordStartupErrorIgnoresNil(t *testing.T) {
	saved := startupErrors
	t.Cleanup(func() { startupErrors = saved })
	startupErrors = nil

	RecordStartupError(nil)

	require.Empty(t, startupErrors)
}
