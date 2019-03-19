package world

import (
	"testing"

	"github.com/faiface/pixel"
	"github.com/google/uuid"
)

func TestNewGate(t *testing.T) {
	type args struct {
		l pixel.Vec
		s gateStatus
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
				w: NewWorld(768, 1024, nil, 2),
			},
			wantErr: false,
		},
		{
			name: "failure",
			args: args{
				l: pixel.V(100, 100),
				s: GateOpen,
				w: NewWorld(68, 1024, nil, 2),
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := tt.args.w.NewGate(tt.args.l, tt.args.s); (err != nil) != tt.wantErr {
				t.Errorf("NewGate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestGate_Reserve(t *testing.T) {
	type fields struct {
		Location pixel.Vec
		Status   gateStatus
		Reserved bool
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
			},
			wantErr: false,
		},
		{
			name: "gate closed",
			fields: fields{
				Location: pixel.V(100, 100),
				Status:   GateClosed,
				Reserved: false,
			},
			wantErr: true,
		},
		{
			name: "gate reserved",
			fields: fields{
				Location: pixel.V(100, 100),
				Status:   GateOpen,
				Reserved: true,
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			g := &Gate{
				Location: tt.fields.Location,
				Status:   tt.fields.Status,
				Reserved: tt.fields.Reserved,
			}
			if err := g.Reserve(uuid.New()); (err != nil) != tt.wantErr {
				t.Errorf("Gate.Reserve() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestGate_UnReserve(t *testing.T) {
	type fields struct {
		Location pixel.Vec
		Status   gateStatus
		Reserved bool
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
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			g := &Gate{
				Location: tt.fields.Location,
				Status:   tt.fields.Status,
				Reserved: tt.fields.Reserved,
			}
			g.UnReserve()
			if g.Reserved != false {
				t.Errorf("Expected gate %v to be unrserved, but it's not.", g)
			}
		})
	}
}
