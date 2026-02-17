package hexcolor

import (
	"image/color"
	"testing"
)

func TestToRGBA(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		want    color.RGBA
		wantErr bool
	}{
		{
			name:  "6-digit hex color",
			input: "#ff8040",
			want:  color.RGBA{R: 0xff, G: 0x80, B: 0x40, A: 0xff},
		},
		{
			name:  "8-digit hex color with alpha",
			input: "#ff804080",
			want:  color.RGBA{R: 0xff, G: 0x80, B: 0x40, A: 0x80},
		},
		{
			name:  "black",
			input: "#000000",
			want:  color.RGBA{R: 0, G: 0, B: 0, A: 0xff},
		},
		{
			name:  "white",
			input: "#ffffff",
			want:  color.RGBA{R: 0xff, G: 0xff, B: 0xff, A: 0xff},
		},
		{
			name:  "fully transparent",
			input: "#00000000",
			want:  color.RGBA{R: 0, G: 0, B: 0, A: 0},
		},
		{
			name:  "lowercase hex",
			input: "#aabbcc",
			want:  color.RGBA{R: 0xaa, G: 0xbb, B: 0xcc, A: 0xff},
		},
		{
			name:  "uppercase hex",
			input: "#AABBCC",
			want:  color.RGBA{R: 0xaa, G: 0xbb, B: 0xcc, A: 0xff},
		},
		{
			name:  "3-digit shorthand",
			input: "#fff",
			want:  color.RGBA{R: 0xff, G: 0xff, B: 0xff, A: 0xff},
		},
		{
			name:  "3-digit shorthand color",
			input: "#f80",
			want:  color.RGBA{R: 0xff, G: 0x88, B: 0x00, A: 0xff},
		},
		{
			name:  "4-digit shorthand with alpha",
			input: "#ccca",
			want:  color.RGBA{R: 0xcc, G: 0xcc, B: 0xcc, A: 0xaa},
		},
		{
			name:    "missing hash prefix",
			input:   "ff8040",
			wantErr: true,
		},
		{
			name:    "invalid hex characters",
			input:   "#gghhii",
			wantErr: true,
		},
		{
			name:    "too long",
			input:   "#ff80401122",
			wantErr: true,
		},
		{
			name:    "empty string",
			input:   "",
			wantErr: true,
		},
		{
			name:    "only hash",
			input:   "#",
			wantErr: true,
		},
		{
			name:    "2 chars",
			input:   "#ff",
			wantErr: true,
		},
		{
			name:    "5 chars",
			input:   "#abcde",
			wantErr: true,
		},
		{
			name:    "7 chars",
			input:   "#abcdeff",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ToRGBA(tt.input)
			if (err != nil) != tt.wantErr {
				t.Fatalf("ToRGBA(%q) error = %v, wantErr %v", tt.input, err, tt.wantErr)
			}
			if !tt.wantErr && got != tt.want {
				t.Errorf("ToRGBA(%q) = %v, want %v", tt.input, got, tt.want)
			}
		})
	}
}
