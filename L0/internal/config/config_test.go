package config

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestMustLoad(t *testing.T) {
	broker := "localhost:9095"
	t.Setenv("KAFKA_BROKERS", broker)
	cfg := MustLoad("./../../.env")
	require.NotNil(t, cfg, "failed to load config")
	require.Equal(t, broker, cfg.Kafka.Brokers, "failed to load config")
}

func TestLoadEnv(t *testing.T) {
	testcases := []struct {
		name        string
		path        string
		expectedErr bool
	}{
		{
			name:        "valid",
			path:        "../../.env",
			expectedErr: false,
		},
		{
			name:        "reading from environment variables",
			path:        "",
			expectedErr: false,
		},
	}
	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {

			err := LoadEnv(tc.path)
			if tc.expectedErr {
				require.Error(t, err, "expected error, got none")
			} else {
				require.NoError(t, err, tc.name, "expected no error")
			}
		})
	}
}

func TestLoadError(t *testing.T) {
	testcases := []struct {
		name  string
		value string
	}{
		{
			name:  "empty value",
			value: "",
		},
		{
			name:  "invalid value",
			value: "invalid",
		},
	}

	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			t.Setenv("DSN_PORT", tc.value)
			cfg, err := Load()
			require.Error(t, err, "expected error")
			require.Nil(t, cfg, "expected nil")
		})
	}
}
