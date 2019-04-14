package utils

import (
	"testing"

	"github.com/faiface/pixel"
)

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

func TestRectContains(t *testing.T) {
	type args struct {
		r1 pixel.Rect
		r2 pixel.Rect
	}
	tests := []struct {
		name string
		args args
		want bool
	}{
		{
			name: "contains",
			args: args{
				r1: pixel.R(0, 0, 10, 10),
				r2: pixel.R(2, 2, 4, 4),
			},
			want: true,
		},
		{
			name: "does not contain, inside",
			args: args{
				r1: pixel.R(2, 2, 4, 4),
				r2: pixel.R(0, 0, 10, 10),
			},
			want: false,
		},
		{
			name: "intersect",
			args: args{
				r1: pixel.R(2, 2, 4, 4),
				r2: pixel.R(0, 0, 3, 3),
			},
			want: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := RectContains(tt.args.r1, tt.args.r2); got != tt.want {
				t.Errorf("RectContains() = %v, want %v", got, tt.want)
			}
		})
	}
}
