package transcoder

import (
	"fmt"
	"os"
	"os/exec"
	"strings"
)

func ProcessVideo(inputPath string, outputDir string) error {
	
	arr := strings.Split(inputPath, "/")
	nName := arr[len(arr)-1]

	fileName := strings.Split(nName, ".")[0]

	videoOutputDir := fmt.Sprintf("%s/%s", outputDir, fileName)

	err := os.MkdirAll(videoOutputDir, os.ModePerm)
	if err != nil {
		return err
	}

	outputPath := fmt.Sprintf("%s/index.m3u8", videoOutputDir)

	// fmt.Println("input path : ", inputPath, fileName)
	// fmt.Println("output path :", outputPath)
	//execute the command
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

	return cmd.Run()
	// return err

}
