package utils

import "testing"

func TestRandomInt(t *testing.T) {
	type args struct {
		min int
		max int
	}
	tests := []struct {
		name    string
		args    args
		succeed bool
	}{
		{
			name:    "success",
			args:    args{0, 10},
			succeed: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := RandomInt(tt.args.min, tt.args.max)
			if tt.succeed {
				if got < tt.args.min || got >= tt.args.max {
					t.Errorf("expected value between %v and %v; got %v", tt.args.min, tt.args.max, got)
				}
			}
		})
	}
}

func TestRandomFloat64(t *testing.T) {
	type args struct {
		min float64
		max float64
	}
	tests := []struct {
		name    string
		args    args
		succeed bool
	}{
		{
			name:    "success",
			args:    args{0.0, 10.0},
			succeed: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := RandomFloat64(tt.args.min, tt.args.max)
			if got < tt.args.min || got >= tt.args.max {
				t.Errorf("expected value between %v and %v; got %v", tt.args.min, tt.args.max, got)
			}
		})
	}
}
