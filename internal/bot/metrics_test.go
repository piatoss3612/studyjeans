package bot

import "testing"

func TestNewBotMetricsServer(t *testing.T) {
	tests := []struct {
		name   string
		port   string
		hasErr bool
		expErr string
	}{
		{
			name:   "valid port",
			port:   "8080",
			hasErr: false,
			expErr: "",
		},
		{
			name:   "invalid port",
			port:   "65536",
			hasErr: true,
			expErr: "invalid port",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := NewBotMetricsServer(tt.port)
			if tt.hasErr && err == nil {
				t.Errorf("expected error but got nil")
			}
			if !tt.hasErr && err != nil {
				t.Errorf("expected nil but got %v", err)
			}
			if tt.hasErr && err != nil && err.Error() != tt.expErr {
				t.Errorf("expected %v but got %v", tt.expErr, err)
			}
		})
	}
}
