package models

import "time"

type VideoProbe struct {
	ID      string `db:"id"`
	VideoID string `db:"video_id"`

	// Container / format
	FormatName     *string  `db:"format_name"`
	FormatLongName *string  `db:"format_long_name"`
	FormatBitrate  *int64   `db:"format_bitrate"`
	FormatDuration *float64 `db:"format_duration"`

	// Video
	VideoCodec         *string `db:"video_codec"`
	VideoCodecLongName *string `db:"video_codec_long_name"`
	VideoProfile       *string `db:"video_profile"`
	VideoLevel         *int    `db:"video_level"`

	Width  *int `db:"width"`
	Height *int `db:"height"`

	VideoBitrate *int64 `db:"video_bitrate"`

	FrameRateNum *int `db:"frame_rate_num"`
	FrameRateDen *int `db:"frame_rate_den"`

	PixelFormat *string `db:"pixel_format"`

	ColorSpace     *string `db:"color_space"`
	ColorTransfer  *string `db:"color_transfer"`
	ColorPrimaries *string `db:"color_primaries"`

	SampleAspectRatio  *string `db:"sample_aspect_ratio"`
	DisplayAspectRatio *string `db:"display_aspect_ratio"`

	// Audio
	AudioCodec         *string `db:"audio_codec"`
	AudioCodecLongName *string `db:"audio_codec_long_name"`

	AudioBitrate *int64 `db:"audio_bitrate"`

	AudioSampleRate    *int    `db:"audio_sample_rate"`
	AudioChannels      *int    `db:"audio_channels"`
	AudioChannelLayout *string `db:"audio_channel_layout"`

	// Probe metadata
	ProbedAt  time.Time `db:"probed_at"`
	CreatedAt time.Time `db:"created_at"`
}
type VideoRendition struct {
	ID           string
	VideoID      string
	Width        int
	Height       int
	Bitrate      int
	FrameRateNum *int
	FrameRateDen *int
}
