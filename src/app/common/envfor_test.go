package common_test

import (
	"testing"

	"github.com/newrelic/release-toolkit/src/app/common"
)

func TestEnvFor(t *testing.T) {
	t.Parallel()

	for _, tc := range []struct {
		name     string
		flag     string
		prefixes []string
		expected string
	}{
		{
			name:     "Simple_Flag",
			flag:     "config",
			expected: "CONFIG",
		},
		{
			name:     "Flag_With_Dashes",
			flag:     "config-file",
			expected: "CONFIG_FILE",
		},
		{
			name:     "Flag_With_Prefix",
			flag:     "config",
			prefixes: []string{"myapp"},
			expected: "MYAPP_CONFIG",
		},
		{
			name:     "Flag_With_Prefix_And_Dashes",
			flag:     "config-file",
			prefixes: []string{"myapp", "command"},
			expected: "MYAPP_COMMAND_CONFIG_FILE",
		},
	} {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			actual := common.EnvFor(tc.flag, tc.prefixes...)
			if tc.expected != actual[0] {
				t.Fatalf("Expected env var to be %s, got %s", tc.expected, actual)
			}
		})
	}
}
