package transcoder

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"os/exec"
	"strings"
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