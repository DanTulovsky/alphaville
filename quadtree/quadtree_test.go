package quadtree

import (
	"testing"

	"github.com/faiface/pixel"
	"github.com/go-test/deep"
	"github.com/stretchr/testify/assert"
	"gogs.wetsnow.com/dant/alphaville/utils"
)

func TestNewTree(t *testing.T) {
	type fields struct {
		bounds pixel.Rect
		level  int
	}
	tests := []struct {
		name   string
		fields fields
		want   *Tree
	}{
		{
			name: "one",
			fields: fields{
				bounds: pixel.R(0, 0, 10, 10),
				level:  0,
			},
			want: &Tree{
				Bounds:  pixel.R(0, 0, 10, 10),
				Level:   0,
				Objects: []pixel.Rect{},
				Nodes:   []*Tree{},
			},
		},
		{
			name: "two",
			fields: fields{
				bounds: pixel.R(70, 56, 30, 10),
				level:  4,
			},
			want: &Tree{
				Bounds:  pixel.R(30, 10, 70, 56),
				Level:   4,
				Objects: []pixel.Rect{},
				Nodes:   []*Tree{},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := NewTree(tt.fields.bounds, tt.fields.level)
			diff := deep.Equal(got, tt.want)
			if len(diff) != 0 {
				t.Errorf("NewTree() = %v, want %v\n  Diff: %v", got, tt.want, diff)
			}
		})
	}
}

func TestTree_split(t *testing.T) {
	type fields struct {
		Bounds  pixel.Rect
		Level   int
		Objects []pixel.Rect
		Nodes   []*Tree
	}
	tests := []struct {
		name   string
		fields fields
	}{
		{
			name: "already split",
			fields: fields{
				Bounds:  pixel.R(0, 0, 100, 100),
				Level:   0,
				Objects: []pixel.Rect{}, // no objects anyway
				Nodes: []*Tree{
					NewTree(pixel.R(0, 50, 50, 100), 1),   // top left
					NewTree(pixel.R(50, 50, 100, 100), 1), // top right
					NewTree(pixel.R(0, 0, 50, 50), 1),     // bottom left
					NewTree(pixel.R(50, 0, 100, 50), 1),   // bottom right
				},
			},
		},
		{
			name: "not split",
			fields: fields{
				Bounds:  pixel.R(0, 0, 100, 100),
				Level:   0,
				Objects: []pixel.Rect{}, // no objects anyway
				Nodes:   []*Tree{},
			},
		},
	}

	assert := assert.New(t)

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			qt := &Tree{
				Bounds:  tt.fields.Bounds,
				Level:   tt.fields.Level,
				Objects: tt.fields.Objects,
				Nodes:   tt.fields.Nodes,
			}
			qt.split()
			assert.Equal(len(qt.Nodes), 4)

			// check sizes
			for _, n := range qt.Nodes {
				assert.Equal(qt.Bounds.Area()/4, n.Bounds.Area())
			}

			// make sure no overlap
			for _, n := range qt.Nodes {
				for _, other := range qt.Nodes {
					if n == other {
						continue // skip yourself
					}
					assert.False(utils.Intersect(n.Bounds, other.Bounds))
				}
			}
		})
	}
}
