package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"math"
	"os"
	"os/exec"
)

func (cfg apiConfig) ensureAssetsDir() error {
	if _, err := os.Stat(cfg.assetsRoot); os.IsNotExist(err) {
		return os.Mkdir(cfg.assetsRoot, 0755)
	}
	return nil
}

func abs(x float64) float64 { return math.Abs(x) }

func getVideoAspectRatio(filePath string) (string, error) {
	var buf bytes.Buffer
	cmd := exec.Command("ffprobe", "-v", "error", "-print_format", "json", "-show_streams", filePath)
	cmd.Stdout = &buf
	if err := cmd.Run(); err != nil {
		return "", fmt.Errorf("ffprobe error: %w", err)
	}

	var result struct {
		Streams []struct {
			Width  int `json:"width"`
			Height int `json:"height"`
		} `json:"streams"`
	}
	if err := json.Unmarshal(buf.Bytes(), &result); err != nil {
		return "", fmt.Errorf("couldn't parse ffprobe output: %w", err)
	}
	if len(result.Streams) == 0 {
		return "", fmt.Errorf("no streams found in video")
	}

	width := result.Streams[0].Width
	height := result.Streams[0].Height

	ratio := float64(width) / float64(height)
	landscape := 16.0 / 9.0
	portrait := 9.0 / 16.0

	threshold := 0.05
	switch {
	case abs(ratio-landscape) < threshold:
		return "16:9", nil
	case abs(ratio-portrait) < threshold:
		return "9:16", nil
	default:
		return "other", nil
	}
}

func processVideoForFastStart(filePath string) (string, error) {
	outputPath := filePath + ".processing"
	cmd := exec.Command("ffmpeg", "-i", filePath, "-c", "copy", "-movflags", "faststart", "-f", "mp4", outputPath)
	if err := cmd.Run(); err != nil {
		return "", fmt.Errorf("ffmpeg error: %w", err)
	}
	return outputPath, nil
}
