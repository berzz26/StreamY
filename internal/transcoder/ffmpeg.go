package transcoder

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"

	"strings"

	"github.com/berzz26/StreamY/internal/models"
)

func ProcessVideo(inputPath string, outputDir string) error {

	err := os.MkdirAll(outputDir, os.ModePerm)
	if err != nil {
		return err
	}

	outputPath := fmt.Sprintf(
		"%s/index.m3u8",
		outputDir,
	)

	cmd := exec.Command(
		"ffmpeg",
		"-hwaccel", "cuda",
		"-hwaccel_output_format", "cuda",
		"-i", inputPath,
		"-preset", "superfast",
		"-g", "48",
		"-sc_threshold", "0",

		"-map", "0:v:0",
		"-map", "0:a:0?",

		"-c:v", "h264_nvenc",
		"-preset", "p5",
		"-c:a", "aac",

		"-f", "hls",
		"-hls_time", "6",
		"-hls_playlist_type", "vod",

		outputPath,
	)

	var stderr bytes.Buffer

	// Keep showing FFmpeg output in the terminal,
	// while also capturing it for the returned error.
	cmd.Stdout = os.Stdout
	cmd.Stderr = io.MultiWriter(os.Stderr, &stderr)

	err = cmd.Run()

	if err != nil {
		ffmpegErr := strings.TrimSpace(stderr.String())

		return fmt.Errorf("ffmpeg: %s", ffmpegErr)
	}

	return nil
}

func EncodeRendition(rendition *models.VideoRendition, outputDir string, inputPath string) error {

	if rendition == nil {
		return fmt.Errorf("rendition is nil")
	}
	resolutionDir := fmt.Sprintf("%dp", rendition.Height)

	newOutputDir := filepath.Join(outputDir, resolutionDir)

	if err := os.MkdirAll(newOutputDir, os.ModePerm); err != nil {
		return err
	}

	outputPath := filepath.Join(newOutputDir, "index.m3u8")

	cmd := exec.Command(
		"ffmpeg",
		"-hwaccel", "cuda",
		"-hwaccel_output_format", "cuda",

		"-i", inputPath,

		"-vf", fmt.Sprintf(
			"scale_cuda=%d:%d",
			rendition.Width,
			rendition.Height,
		),

		"-map", "0:v:0",
		"-map", "0:a:0?",

		"-c:v", "h264_nvenc",
		"-preset", "p4/fast",
		"-b:v", fmt.Sprintf("%d", rendition.Bitrate),

		"-c:a", "aac",
		"-b:a", "128k",

		"-g", "48",
		"-sc_threshold", "0",

		"-f", "hls",
		"-hls_time", "6",
		"-hls_playlist_type", "vod",

		"-hls_segment_filename",
		filepath.Join(newOutputDir, "segment_%03d.ts"),

		outputPath,
	)
	var stderr bytes.Buffer

	// Keep showing FFmpeg output in the terminal,
	// while also capturing it for the returned error.
	cmd.Stdout = os.Stdout
	cmd.Stderr = io.MultiWriter(os.Stderr, &stderr)

	err := cmd.Run()

	if err != nil {
		ffmpegErr := strings.TrimSpace(stderr.String())

		return fmt.Errorf("ffmpeg: %s", ffmpegErr)
	}

	return nil

}
