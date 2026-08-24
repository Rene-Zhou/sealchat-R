package api

import (
	"testing"

	"sealchat/model"
	"sealchat/protocol"
)

func TestNormalizeStatePayloadPlacementCompatibility(t *testing.T) {
	form := &model.ChannelIFormModel{
		DefaultWidth:  640,
		DefaultHeight: 360,
	}
	tests := []struct {
		name          string
		state         protocol.ChannelIFormStatePayload
		wantPlacement string
		wantFloating  bool
	}{
		{name: "legacy top", state: protocol.ChannelIFormStatePayload{}, wantPlacement: "top"},
		{name: "legacy floating", state: protocol.ChannelIFormStatePayload{Floating: true}, wantPlacement: "floating", wantFloating: true},
		{name: "explicit top wins", state: protocol.ChannelIFormStatePayload{Placement: "top", Floating: true}, wantPlacement: "top"},
		{name: "explicit right wins", state: protocol.ChannelIFormStatePayload{Placement: "right", Floating: true}, wantPlacement: "right"},
		{name: "explicit floating wins", state: protocol.ChannelIFormStatePayload{Placement: "floating"}, wantPlacement: "floating", wantFloating: true},
		{name: "invalid falls back", state: protocol.ChannelIFormStatePayload{Placement: "side"}, wantPlacement: "top"},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			normalized := normalizeStatePayload(test.state, form, false)
			if normalized.Placement != test.wantPlacement {
				t.Fatalf("placement = %q, want %q", normalized.Placement, test.wantPlacement)
			}
			if normalized.Floating != test.wantFloating {
				t.Fatalf("floating = %v, want %v", normalized.Floating, test.wantFloating)
			}
			if normalized.Width != form.DefaultWidth || normalized.Height != form.DefaultHeight {
				t.Fatalf("size = %dx%d, want %dx%d", normalized.Width, normalized.Height, form.DefaultWidth, form.DefaultHeight)
			}
		})
	}
}
