package driver

import (
	"crypto/sha256"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// IIASA CUSTOM: Support shared FUSE mount pods for read-only overlay mounts

// IsSharedMount checks if the volume mount is ReadOnly or has the overlay flag, and the accelerator is enabled.
func IsSharedMount(sourceType string, opts MountOptions) bool {
	if os.Getenv("ACCELERATOR_MOUNT") == "" {
		return false
	}
	if opts.ReadOnly {
		return true
	}
	for _, arg := range opts.ExtraArgs {
		if arg == "--overlay" || arg == "overlay" {
			return true
		}
	}
	return false
}

// SharedVolumeID generates a deterministic shared volume ID for read-only overlay mounts.
func SharedVolumeID(sourceType, sourceID string, opts MountOptions) string {
	key := fmt.Sprintf("%s-%s", sourceType, sourceID)
	if sourceType == "repo" && opts.Revision != "" {
		key = fmt.Sprintf("%s-%s-%s", sourceType, sourceID, opts.Revision)
	}
	h := sha256.Sum256([]byte(key))
	return fmt.Sprintf("shared-%x", h[:6])
}

// ResolveTokenFilePath returns the correct token file path, using the shared volume ID if applicable.
func ResolveTokenFilePath(cacheBase, volumeID, sourceType, sourceID string, opts MountOptions) string {
	tokVolID := volumeID
	if IsSharedMount(sourceType, opts) {
		tokVolID = SharedVolumeID(sourceType, sourceID, opts)
	}
	return tokenFilePath(cacheBase, tokVolID)
}

// ResolveCacheDir returns the correct cache directory, using the shared volume ID if applicable.
func ResolveCacheDir(cacheBase, volumeID, sourceType, sourceID string, opts MountOptions, customCacheDir string) string {
	if customCacheDir != "" {
		return customCacheDir
	}
	cacheVolID := volumeID
	if IsSharedMount(sourceType, opts) {
		cacheVolID = SharedVolumeID(sourceType, sourceID, opts)
	}
	return filepath.Join(cacheBase, sanitizeVolumeID(cacheVolID))
}

// CleanSharedToken cleans up the token file for shared mounts when the mount pod is deleted.
func CleanSharedToken(cacheDir, volumeID string) {
	if strings.HasPrefix(volumeID, "shared-") {
		tokenFile := filepath.Join(cacheDir, volumeID, "token")
		_ = os.Remove(tokenFile)
		_ = os.Remove(filepath.Dir(tokenFile))
	}
}
