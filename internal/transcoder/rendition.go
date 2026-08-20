package transcoder

import (
	"github.com/google/uuid"
	"github.com/berzz26/StreamY/internal/models"
	"math"
)

var resolutionLadder = []int{
	2160,
	1440,
	1080,
	720,
	480,
	360,
	240,
}

var bitrateLadder = map[int]int{
	2160: 12000000,
	1440: 8000000,
	1080: 5000000,
	720:  2500000,
	480:  1000000,
	360:  600000,
	240:  400000,
}

func PlanRenditions(
	probe *models.VideoProbe,
) []models.VideoRendition {
	var renditions []models.VideoRendition

	if probe.Width == nil || probe.Height == nil {
		return renditions
	}

	sourceWidth := *probe.Width
	sourceHeight := *probe.Height

	for _, height := range resolutionLadder {
		//never upscale
		if sourceHeight < height {
			continue
		}
		// maintaining the aspect ratio while calculating the new width
		width := int(
			math.Round(
				float64(sourceWidth) * float64(height) / float64(sourceHeight),
			),
		)

		// keeping the width as even (so 1281 gets to 1280)
		width = width - (width % 2)

		rendition := models.VideoRendition{
			ID:           uuid.New().String(),
			VideoID:      probe.VideoID,
			Width:        width,
			Height:       height,
			Bitrate:      bitrateLadder[height],
			FrameRateNum: probe.FrameRateNum,
			FrameRateDen: probe.FrameRateDen,
		}

		renditions = append(renditions, rendition)
		
	}

	return renditions
}
