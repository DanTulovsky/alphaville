package world

import (
	"testing"

	"github.com/faiface/pixel"
)

func TestHaveCollisions(t *testing.T) {
	type args struct {
		r1 pixel.Rect
		r2 pixel.Rect
		v1 pixel.Vec
		v2 pixel.Vec
	}
	tests := []struct {
		name string
		args args
		want bool
	}{
		{
			name: "no collision",
			args: args{
				r1: pixel.R(0, 0, 10, 10),
				r2: pixel.R(20, 20, 40, 40),
				v1: pixel.V(0, 0),
				v2: pixel.V(0, 0),
			},
			want: false,
		},
		{
			name: "collision pass through",
			args: args{
				r1: pixel.R(0, 0, 10, 10),
				r2: pixel.R(11, 0, 21, 20),
				v1: pixel.V(50, 0),
				v2: pixel.V(0, 0),
			},
			want: true,
		},
		{
			name: "collision intersect",
			args: args{
				r1: pixel.R(0, 0, 10, 10),
				r2: pixel.R(11, 0, 21, 20),
				v1: pixel.V(5, 0),
				v2: pixel.V(0, 0),
			},
			want: true,
		},
		{
			name: "collision pass through",
			args: args{
				r1: pixel.R(0, 0, 10, 10),
				r2: pixel.R(11, 0, 21, 20),
				v1: pixel.V(50, 1),
				v2: pixel.V(0, 0),
			},
			want: true,
		},
		{
			name: "collision intersect 2",
			args: args{
				r1: pixel.R(0, 0, 10, 10),
				r2: pixel.R(11, 0, 21, 20),
				v1: pixel.V(5, 0),
				v2: pixel.V(1, 0),
			},
			want: true,
		},
		{
			name: "collision pass through 2",
			args: args{
				r1: pixel.R(0, 0, 10, 10),
				r2: pixel.R(11, 0, 21, 20),
				v1: pixel.V(50, 1),
				v2: pixel.V(-20, 0),
			},
			want: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := HaveCollisions(tt.args.r1, tt.args.r2, tt.args.v1, tt.args.v2); got != tt.want {
				t.Errorf("[%v] HaveCollisions() = %v, want %v", tt.name, got, tt.want)
			}
		})
	}
}
