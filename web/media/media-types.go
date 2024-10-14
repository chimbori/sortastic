package media

import (
	"path/filepath"
	"slices"
	"strings"
)

var (
	imageExts   = []string{".png", ".jpg", ".jpeg", ".gif", ".webp"}
	videoExts   = []string{".mp4", ".mov", ".webm"}
	sidecarExts = []string{".dng", ".xmp"}
)

// MIME Types
const (
	Unknown = ""
	Image   = "image"
	Video   = "video"
)

func getMediaType(filename string) string {
	ext := strings.ToLower(filepath.Ext(filename))
	if slices.Contains(imageExts, ext) {
		return Image
	} else if slices.Contains(videoExts, ext) {
		return Video
	}
	return Unknown
}

func getSidecarFiles(filename string) []string {
	basename := filename[:len(filename)-len(filepath.Ext(filename))]
	var sidecars []string
	for _, sidecarExt := range sidecarExts {
		sidecars = append(sidecars, basename+sidecarExt)
	}
	return sidecars
}
