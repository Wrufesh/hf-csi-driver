package driver

import (
	"os"
	"testing"
)

func TestIsSharedMount(t *testing.T) {
	// Clean up env at the end of the test
	defer os.Unsetenv("ACCELERATOR_MOUNT")

	tests := []struct {
		name       string
		envVal     string
		sourceType string
		readOnly   bool
		extraArgs  []string
		want       bool
	}{
		{
			name:       "Accelerator disabled",
			envVal:     "",
			sourceType: "bucket",
			readOnly:   true,
			extraArgs:  []string{"--overlay"},
			want:       false,
		},
		{
			name:       "ReadOnly without overlay",
			envVal:     "1",
			sourceType: "bucket",
			readOnly:   true,
			extraArgs:  nil,
			want:       true,
		},
		{
			name:       "ReadWrite with overlay (--overlay)",
			envVal:     "1",
			sourceType: "bucket",
			readOnly:   false,
			extraArgs:  []string{"--overlay"},
			want:       true,
		},
		{
			name:       "ReadWrite with overlay (overlay)",
			envVal:     "1",
			sourceType: "bucket",
			readOnly:   false,
			extraArgs:  []string{"overlay"},
			want:       true,
		},
		{
			name:       "ReadWrite without overlay",
			envVal:     "1",
			sourceType: "bucket",
			readOnly:   false,
			extraArgs:  []string{"advanced-writes"},
			want:       false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.envVal != "" {
				os.Setenv("ACCELERATOR_MOUNT", tt.envVal)
			} else {
				os.Unsetenv("ACCELERATOR_MOUNT")
			}

			opts := MountOptions{
				ReadOnly:  tt.readOnly,
				ExtraArgs: tt.extraArgs,
			}

			got := IsSharedMount(tt.sourceType, opts)
			if got != tt.want {
				t.Errorf("IsSharedMount() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestSharedVolumeID(t *testing.T) {
	opts := MountOptions{
		Revision: "main",
	}
	nodeID := "test-node-1"

	id1 := SharedVolumeID(nodeID, "bucket", "my-project", opts)
	id2 := SharedVolumeID(nodeID, "bucket", "my-project", opts)
	if id1 != id2 {
		t.Errorf("SharedVolumeID should be deterministic: %q != %q", id1, id2)
	}

	id3 := SharedVolumeID(nodeID, "bucket", "another-project", opts)
	if id1 == id3 {
		t.Errorf("SharedVolumeID should differ for different source IDs: %q == %q", id1, id3)
	}
}
