package world

import (
	"testing"
	"time"

	"golang.org/x/image/colornames"

	"github.com/faiface/pixel"
)

func TestNewGate(t *testing.T) {
	type args struct {
		l pixel.Vec
		s GateStatus
		c time.Duration
		r float64
		w *World
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name: "success",
			args: args{
				l: pixel.V(10, 10),
				s: GateOpen,
				w: NewWorld(768, 1024, nil, 2, 2, &DebugConfig{}),
				c: time.Minute * 1,
				r: 20,
			},
			wantErr: false,
		},
		{
			name: "failure",
			args: args{
				l: pixel.V(100, 100),
				s: GateOpen,
				w: NewWorld(68, 1024, nil, 2, 2, &DebugConfig{}),
				c: time.Minute * 1,
				r: 20,
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			g := NewGate("", tt.args.l, tt.args.s, tt.args.c, tt.args.r)
			if err := tt.args.w.AddGate(g); (err != nil) != tt.wantErr {
				t.Errorf("NewGate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestGate_Reserve(t *testing.T) {
	type fields struct {
		Location pixel.Vec
		Status   GateStatus
		Reserved bool
		coolDown time.Duration
		object   BaseObject
	}
	tests := []struct {
		name    string
		fields  fields
		wantErr bool
	}{
		{
			name: "success",
			fields: fields{
				Location: pixel.V(100, 100),
				Status:   GateOpen,
				Reserved: false,
				coolDown: time.Minute * 1,
				object:   NewBaseObject("", colornames.Red, 0, 0),
			},
			wantErr: false,
		},
		{
			name: "gate closed",
			fields: fields{
				Location: pixel.V(100, 100),
				Status:   GateClosed,
				Reserved: false,
				coolDown: time.Minute * 1,
				object:   NewBaseObject("", colornames.Red, 0, 0),
			},
			wantErr: true,
		},
		{
			name: "gate reserved",
			fields: fields{
				Location: pixel.V(100, 100),
				Status:   GateOpen,
				Reserved: true,
				coolDown: time.Minute * 1,
				object:   NewBaseObject("", colornames.Red, 0, 0),
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			g := NewGate("", tt.fields.Location, tt.fields.Status, tt.fields.coolDown, 10)
			if tt.fields.Reserved {
				if err := g.Reserve(&tt.fields.object); err != nil {
					t.Errorf("failed to reserve gate: %v", err)
				}
			}
			if err := g.Reserve(&tt.fields.object); (err != nil) != tt.wantErr {
				t.Errorf("Gate.Reserve() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestGate_UnReserve(t *testing.T) {
	type fields struct {
		Location pixel.Vec
		Status   GateStatus
		Reserved bool
		object   BaseObject
	}
	tests := []struct {
		name   string
		fields fields
	}{
		{
			name: "success",
			fields: fields{
				Location: pixel.V(100, 100),
				Status:   GateOpen,
				Reserved: false,
				object:   NewBaseObject("", colornames.Red, 0, 0),
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			g := NewGate("", tt.fields.Location, tt.fields.Status, 0, 10)
			if tt.fields.Reserved {
				if err := g.Reserve(&tt.fields.object); err != nil {
					t.Errorf("failed to reserve gate: %v", err)
				}
			}
			g.Release()
			if g.Reserved != false {
				t.Errorf("Expected gate %v to be unreserved, but it's not.", g)
			}
		})
	}
}
