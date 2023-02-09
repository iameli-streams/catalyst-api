package video

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestGetPlaybackProfiles(t *testing.T) {
	tests := []struct {
		name  string
		track InputTrack
		want  []EncodedProfile
	}{
		{
			name: "360p input",
			track: InputTrack{
				Type:    "video",
				Bitrate: 1_000_001,
				VideoTrack: VideoTrack{
					Width:  640,
					Height: 360,
				},
			},
			want: []EncodedProfile{
				{Name: "low-bitrate", Width: 640, Height: 360, Bitrate: 500_000},
				{Name: "360p0", Width: 640, Height: 360, Bitrate: 1_000_001},
			},
		},
		{
			name: "low bitrate 360p input",
			track: InputTrack{
				Type:    "video",
				Bitrate: 500_000,
				VideoTrack: VideoTrack{
					Width:  640,
					Height: 360,
				},
			},
			want: []EncodedProfile{
				{Name: "low-bitrate", Width: 640, Height: 360, Bitrate: 250_000},
				{Name: "360p0", Width: 640, Height: 360, Bitrate: 500_000},
			},
		},
		{
			name: "720p input",
			track: InputTrack{
				Type:    "video",
				Bitrate: 4_000_001,
				VideoTrack: VideoTrack{
					Width:  1280,
					Height: 720,
				},
			},
			want: []EncodedProfile{
				{Name: "360p0", Width: 640, Height: 360, Bitrate: 1_000_000},
				{Name: "720p0", Width: 1280, Height: 720, Bitrate: 4_000_001},
			},
		},
		{
			name: "low bitrate 720p input",
			track: InputTrack{
				Type:    "video",
				Bitrate: 1_000_001,
				VideoTrack: VideoTrack{
					Width:  1200,
					Height: 720,
				},
			},
			want: []EncodedProfile{
				{Name: "360p0", Width: 640, Height: 360, Bitrate: 1_000_000},
				{Name: "720p0", Width: 1200, Height: 720, Bitrate: 1_000_001},
			},
		},
		{
			name: "1080p input",
			track: InputTrack{
				Type:    "video",
				Bitrate: 5_000_000,
				VideoTrack: VideoTrack{
					Width:  1920,
					Height: 1080,
				},
			},
			want: []EncodedProfile{
				{Name: "360p0", Width: 640, Height: 360, Bitrate: 1_000_000},
				{Name: "720p0", Width: 1280, Height: 720, Bitrate: 4_000_000},
				{Name: "1080p0", Width: 1920, Height: 1080, Bitrate: 5_000_000},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := GetPlaybackProfiles(InputVideo{
				Tracks: []InputTrack{tt.track},
			})
			require.NoError(t, err)
			require.Equal(t, tt.want, got)
		})
	}
}