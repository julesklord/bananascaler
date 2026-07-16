package tui

import "testing"

func TestParseStep(t *testing.T) {
	tests := []struct {
		msg           string
		expectedStage int
		expectedName  string
	}{
		{
			msg:           "[1/3] Extracting frames...",
			expectedStage: 1,
			expectedName:  "Extracting frames...",
		},
		{
			msg:           "[2/3] Neural upscaling...",
			expectedStage: 2,
			expectedName:  "Neural upscaling...",
		},
		{
			msg:           "[3/3] Re-encoding and muxing...",
			expectedStage: 3,
			expectedName:  "Re-encoding and muxing...",
		},
		{
			msg:           "No brackets here",
			expectedStage: 0,
			expectedName:  "",
		},
		{
			msg:           "[] Empty brackets",
			expectedStage: 0,
			expectedName:  "Empty brackets",
		},
		{
			msg:           "[12/3] Custom stage digits",
			expectedStage: 12,
			expectedName:  "Custom stage digits",
		},
	}

	for _, tc := range tests {
		t.Run(tc.msg, func(t *testing.T) {
			stage, name := parseStep(tc.msg)
			if stage != tc.expectedStage {
				t.Errorf("expected stage %d, got %d", tc.expectedStage, stage)
			}
			if name != tc.expectedName {
				t.Errorf("expected name %q, got %q", tc.expectedName, name)
			}
		})
	}
}
