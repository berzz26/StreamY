package transcoder

import (
	"fmt"
	"os"
	"os/exec"
	"time"
	"log"
	// "strings"
)

func ProcessVideo(inputPath string, outputDir string) error {

	// arr := strings.Split(inputPath, "/")
	// nName := arr[len(arr)-1]

	// fileName := strings.Split(nName, ".")[0]

	start := time.Now()

	err := os.MkdirAll(outputDir, os.ModePerm)
	if err != nil {
		return err
	}

	outputPath := fmt.Sprintf(
		"%s/index.m3u8",
		outputDir,
	)
	// fmt.Println("input path : ", inputPath, fileName)
	// fmt.Println("output path :", outputPath)
	//execute the command
	cmd := exec.Command(
		"ffmpeg",
		"-hwaccel", "cuda",
		"-hwaccel_output_format", "cuda",
		"-i", inputPath,
		"-preset", "fast",
		"-g", "48",
		"-sc_threshold", "0",

		"-map", "0:v:0",
		"-map", "0:a:0",

		"-c:v", "h264_nvenc",
		"-preset", "p5", // p1 fastest, p7 highest quality
		"-c:a", "aac",

		"-f", "hls",

		"-hls_time", "6",

		"-hls_playlist_type", "vod",
		outputPath,
	)

	elapsed := time.Since(start)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	log.Printf("Transcoding completed, transcode_time=%s\n", elapsed)

	return cmd.Run()
	// return err

}
