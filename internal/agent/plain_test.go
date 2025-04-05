package agent

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestPlainParser_Parse(t *testing.T) {
	tests := []struct {
		name    string
		pattern string
		line    string
		data    map[string]string
		want    map[string]string
		wantErr bool
	}{
		{
			name:    "parse plain log",
			pattern: `^(?P<time>[^ ]+ [^ ]+) (?P<level>\w+) (?P<message>.*)`,
			line:    "2023-01-01 12:00:00 INFO Test message",
			data:    nil,
			want: map[string]string{
				"time":    "2023-01-01 12:00:00",
				"level":   "INFO",
				"message": "Test message",
			},
			wantErr: false,
		},
		{
			name:    "non-matching regex",
			pattern: `(?P<time>[^ ]+ [^ ]+) (?P<level>\w+) (?P<message>.*)`,
			line:    "Test message",
			data:    nil,
			want:    nil,
			wantErr: false,
		},
		{
			name:    "partial mathching regex",
			pattern: `^(?P<time>[^ ]+ [^ ]+) (?P<level>\w+)`,
			line:    "2023-01-01 12:00:00 INFO Test message",
			data:    nil,
			want: map[string]string{
				"time":  "2023-01-01 12:00:00",
				"level": "INFO",
			},
			wantErr: false,
		},
		{
			name:    "complex log format",
			pattern: `\[(?P<timestamp>[^\]]+)\] \[(?P<level>[^\]]+)\] \[(?P<module>[^\]]+)\] (?P<message>.*)`,
			line:    "[2023-01-01 12:00:00] [INFO] [auth] User login successful",
			data:    nil,
			want: map[string]string{
				"timestamp": "2023-01-01 12:00:00",
				"level":     "INFO",
				"module":    "auth",
				"message":   "User login successful",
			},
			wantErr: false,
		},
		{
			name:    "join external data",
			pattern: `^(?P<time>[^ ]+ [^ ]+) (?P<level>\w+) (?P<message>.*)`,
			line:    "2023-01-01 12:00:00 INFO Test message",
			data: map[string]string{
				"source": "test_source",
				"host":   "test_host",
			},
			want: map[string]string{
				"time":    "2023-01-01 12:00:00",
				"level":   "INFO",
				"message": "Test message",
				"source":  "test_source",
				"host":    "test_host",
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			parser, err := newPlainParser(tt.pattern)
			require.NoError(t, err, "Failed to initialize FileAgent")
			got, err := parser.Parse(tt.line, tt.data)
			if tt.wantErr {
				require.Error(t, err, "Expected error but got none")
			} else {
				require.NoError(t, err, "Unexpected error")
				require.Equal(t, tt.want, got, "Map not equal", tt.name)
			}
		})
	}
}
