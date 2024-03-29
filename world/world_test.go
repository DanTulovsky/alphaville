package world

import (
	"testing"

	"github.com/go-test/deep"
)

func TestNewWorld(t *testing.T) {
	type args struct {
		x        float64
		y        float64
		ground   Object
		gravity  float64
		maxSpeed float64
	}
	tests := []struct {
		name string
		args args
		want *World
	}{
		{
			name: "success",
			args: args{
				x:        100,
				y:        200,
				ground:   nil,
				gravity:  2,
				maxSpeed: 2,
			},
			want: &World{
				X:              100,
				Y:              200,
				Objects:        []Object{},
				Gates:          []*Gate{},
				Ground:         nil,
				gravity:        2,
				MaxObjectSpeed: 2,
				Stats:          NewStats(nil),
				ManualControl:  NewNullObject(),
				console:        nil,
				MinObjectSide:  20,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := NewWorld(tt.args.x, tt.args.y, tt.args.ground, tt.args.gravity, tt.args.maxSpeed, &DebugConfig{}, nil)
			diff := deep.Equal(got, tt.want)
			if len(diff) != 0 {
				t.Errorf("NewWorld() = %v, want %v\nDiff: %v", got, tt.want, diff)
			}
		})
	}
}
