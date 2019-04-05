package graph

import (
	"testing"

	"github.com/faiface/pixel"
)

func TestLinesIntersect(t *testing.T) {
	type args struct {
		l1 Edge
		l2 Edge
	}
	tests := []struct {
		name string
		args args
		want bool
	}{
		{
			name: "no intersect, parallel",
			args: args{
				l1: Edge{
					A: pixel.V(0, 0),
					B: pixel.V(10, 0),
				},
				l2: Edge{
					A: pixel.V(0, 10),
					B: pixel.V(10, 10),
				},
			},
			want: false,
		},
		{
			name: "no intersect, parallel, vertical",
			args: args{
				l1: Edge{
					A: pixel.V(0, 0),
					B: pixel.V(0, 10),
				},
				l2: Edge{
					A: pixel.V(10, 0),
					B: pixel.V(10, 10),
				},
			},
			want: false,
		},
		{
			name: "intersect, perpendicular",
			args: args{
				l1: Edge{
					A: pixel.V(0, 0),
					B: pixel.V(10, 0),
				},
				l2: Edge{
					A: pixel.V(5, -5),
					B: pixel.V(5, 10),
				},
			},
			want: true,
		},
		{
			name: "l1 covers l2",
			args: args{
				l1: Edge{
					A: pixel.V(2, 10),
					B: pixel.V(20, 10),
				},
				l2: Edge{
					A: pixel.V(10, 10),
					B: pixel.V(30, 10),
				},
			},
			want: true,
		},
		{
			name: "no",
			args: args{
				l1: Edge{
					A: pixel.V(1, 1),
					B: pixel.V(10, 1),
				},
				l2: Edge{
					A: pixel.V(1, 2),
					B: pixel.V(10, 2),
				},
			},
			want: false,
		},
		{
			name: "yes",
			args: args{
				l1: Edge{
					A: pixel.V(10, 0),
					B: pixel.V(0, 10),
				},
				l2: Edge{
					A: pixel.V(0, 0),
					B: pixel.V(10, 10),
				},
			},
			want: true,
		},
		{
			name: "no2",
			args: args{
				l1: Edge{
					A: pixel.V(-5, -5),
					B: pixel.V(0, 0),
				},
				l2: Edge{
					A: pixel.V(1, 1),
					B: pixel.V(10, 10),
				},
			},
			want: false,
		},
		{
			name: "l2covers l1",
			args: args{
				l1: Edge{
					A: pixel.V(20, 10),
					B: pixel.V(40, 10),
				},
				l2: Edge{
					A: pixel.V(10, 10),
					B: pixel.V(30, 10),
				},
			},
			want: true,
		},
		{
			name: "yes",
			args: args{
				l1: Edge{
					A: pixel.V(600, 600),
					B: pixel.V(925, 150),
				},
				l2: Edge{
					A: pixel.V(740, 255),
					B: pixel.V(925, 255),
				},
			},
			want: true,
		},
		{
			name: "yes",
			args: args{
				l1: Edge{
					A: pixel.V(600, 600),
					B: pixel.V(925, 150),
				},
				l2: Edge{
					A: pixel.V(740, 255),
					B: pixel.V(925, 255.24336770882053),
				},
			},
			want: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := EdgesIntersect(tt.args.l1, tt.args.l2); got != tt.want {
				t.Errorf("[%v] LinesIntersect() = %v, want %v", tt.name, got, tt.want)
			}
		})
	}
}

func Test_orientation(t *testing.T) {
	type args struct {
		p pixel.Vec
		q pixel.Vec
		r pixel.Vec
	}
	tests := []struct {
		name string
		args args
		want int
	}{
		{
			name: "above",
			args: args{
				p: pixel.V(0, 0),
				q: pixel.V(10, 0),
				r: pixel.V(5, 10),
			},
			want: 2,
		},
		{
			name: "below",
			args: args{
				p: pixel.V(0, 0),
				q: pixel.V(10, 0),
				r: pixel.V(5, -10),
			},
			want: 1,
		},
		{
			name: "colinear",
			args: args{
				p: pixel.V(0, 0),
				q: pixel.V(10, 0),
				r: pixel.V(20, 0),
			},
			want: 0,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := Orientation(tt.args.p, tt.args.q, tt.args.r); got != tt.want {
				t.Errorf("orientation() = %v, want %v", got, tt.want)
			}
		})
	}
}
