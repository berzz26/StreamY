package transcoder

import (
	"fmt"
	"os"
	"os/exec"
)

func ProcessVideo(inputPath string, outputDir string) error {
	err := os.MkdirAll(outputDir, os.ModePerm) //set perm to the output dir video
	if err != nil {
		return err
	}

	outputPath := fmt.Sprintf("%s/index.m3u8", outputDir)

	cmd := exec.Command(
		"ffmpeg",
		"-i", inputPath,
		"-preset", "fast",
		"-g", "48",
		"-sc_threshold", "0",

		"-map", "0:v:0",
		"-map", "0:a:0",

		"-c:v", "libx264",
		"-c:a", "aac",

		"-f", "hls",

		"-hls_time", "6",

		"-hls_playlist_type", "vod",
		outputPath,
	)

	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	return  cmd.Run()

}
