package transcoder

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os/exec"
	"strconv"
	"strings"
	"time"

	"github.com/berzz26/StreamY/internal/models"
	"github.com/google/uuid"
)

type FFProbeResult struct {
	Streams []FFProbeStream `json:"streams"`
	Format  FFProbeFormat   `json:"format"`
}

type FFProbeStream struct {
	CodecName      string `json:"codec_name"`
	CodecLongName  string `json:"codec_long_name"`
	Profile        string `json:"profile"`
	CodecType      string `json:"codec_type"`
	Width          int    `json:"width"`
	Height         int    `json:"height"`
	Level          int    `json:"level"`
	BitRate        string `json:"bit_rate"`
	RFrameRate     string `json:"r_frame_rate"`
	AvgFrameRate   string `json:"avg_frame_rate"`
	PixelFormat    string `json:"pix_fmt"`
	ColorSpace     string `json:"color_space"`
	ColorTransfer  string `json:"color_transfer"`
	ColorPrimaries string `json:"color_primaries"`

	SampleAspectRatio  string `json:"sample_aspect_ratio"`
	DisplayAspectRatio string `json:"display_aspect_ratio"`

	SampleRate    string `json:"sample_rate"`
	Channels      int    `json:"channels"`
	ChannelLayout string `json:"channel_layout"`
}

type FFProbeFormat struct {
	FormatName     string `json:"format_name"`
	FormatLongName string `json:"format_long_name"`
	BitRate        string `json:"bit_rate"`
	Duration       string `json:"duration"`
}

func ProbeVideo(inputPath string) (*models.VideoProbe, error) {
	cmd := exec.Command(
		"ffprobe",
		"-v", "error",
		"-show_format",
		"-show_streams",
		"-of", "json",
		inputPath,
	)

	var stdout bytes.Buffer
	var stderr bytes.Buffer

	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	if err := cmd.Run(); err != nil {
		return nil, fmt.Errorf(
			"ffprobe failed: %s",
			strings.TrimSpace(stderr.String()),
		)
	}

	var result FFProbeResult

	if err := json.Unmarshal(stdout.Bytes(), &result); err != nil {
		return nil, fmt.Errorf("failed to parse ffprobe output: %w", err)
	}

	probe := &models.VideoProbe{
		ID:        uuid.New().String(),
		ProbedAt:  time.Now(),
		CreatedAt: time.Now(),
	}

	// Format information
	probe.FormatName = stringPtr(result.Format.FormatName)
	probe.FormatLongName = stringPtr(result.Format.FormatLongName)
	probe.FormatBitrate = int64Ptr(parseInt64(result.Format.BitRate))
	probe.FormatDuration = float64Ptr(parseFloat64(result.Format.Duration))

	// Find video/audio streams
	for _, stream := range result.Streams {

		switch stream.CodecType {

		case "video":
			probe.VideoCodec = stringPtr(stream.CodecName)
			probe.VideoCodecLongName = stringPtr(stream.CodecLongName)
			probe.VideoProfile = stringPtr(stream.Profile)
			probe.VideoLevel = intPtr(stream.Level)

			probe.Width = intPtr(stream.Width)
			probe.Height = intPtr(stream.Height)

			probe.VideoBitrate = int64Ptr(parseInt64(stream.BitRate))

			num, den := parseFraction(stream.RFrameRate)
			probe.FrameRateNum = intPtr(num)
			probe.FrameRateDen = intPtr(den)

			probe.PixelFormat = stringPtr(stream.PixelFormat)

			probe.ColorSpace = stringPtr(stream.ColorSpace)
			probe.ColorTransfer = stringPtr(stream.ColorTransfer)
			probe.ColorPrimaries = stringPtr(stream.ColorPrimaries)

			probe.SampleAspectRatio = stringPtr(stream.SampleAspectRatio)
			probe.DisplayAspectRatio = stringPtr(stream.DisplayAspectRatio)

		case "audio":
			probe.AudioCodec = stringPtr(stream.CodecName)
			probe.AudioCodecLongName = stringPtr(stream.CodecLongName)

			probe.AudioBitrate = int64Ptr(parseInt64(stream.BitRate))

			probe.AudioSampleRate = intPtr(
				parseInt(stream.SampleRate),
			)

			probe.AudioChannels = intPtr(stream.Channels)
			probe.AudioChannelLayout = stringPtr(stream.ChannelLayout)
		}
	}

	return probe, nil
}

func parseFraction(value string) (int, int) {
	parts := strings.Split(value, "/")

	if len(parts) != 2 {
		return 0, 0
	}

	num, err1 := strconv.Atoi(parts[0])
	den, err2 := strconv.Atoi(parts[1])

	if err1 != nil || err2 != nil {
		return 0, 0
	}

	return num, den
}

func parseInt(value string) int {
	if value == "" {
		return 0
	}

	result, err := strconv.Atoi(value)
	if err != nil {
		return 0
	}

	return result
}

func parseInt64(value string) int64 {
	if value == "" {
		return 0
	}

	result, err := strconv.ParseInt(value, 10, 64)
	if err != nil {
		return 0
	}

	return result
}

func parseFloat64(value string) float64 {
	if value == "" {
		return 0
	}

	result, err := strconv.ParseFloat(value, 64)
	if err != nil {
		return 0
	}

	return result
}

func stringPtr(value string) *string {
	if value == "" {
		return nil
	}

	return &value
}

func intPtr(value int) *int {
	if value == 0 {
		return nil
	}

	return &value
}

func int64Ptr(value int64) *int64 {
	if value == 0 {
		return nil
	}

	return &value
}

func float64Ptr(value float64) *float64 {
	if value == 0 {
		return nil
	}

	return &value
}
