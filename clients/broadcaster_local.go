package clients

import (
	"encoding/json"
	"fmt"
	"io"
	"net/url"

	"github.com/livepeer/catalyst-api/config"
	"github.com/livepeer/catalyst-api/video"
)

// Currently only implemented by LocalBroadcasterClient
// TODO: Try to come up with a unified interface across Local and Remote
type BroadcasterClient interface {
	TranscodeSegment(segment io.Reader, sequenceNumber int64, profiles []video.EncodedProfile, durationMillis int64, manifestID string) (TranscodeResult, error)
}

type LocalBroadcasterClient struct {
	broadcasterURL url.URL
}

func NewLocalBroadcasterClient(broadcasterURL string) (LocalBroadcasterClient, error) {
	u, err := url.Parse(broadcasterURL)
	if err != nil {
		return LocalBroadcasterClient{}, fmt.Errorf("error parsing local broadcaster URL %q: %s", config.DefaultBroadcasterURL, err)
	}
	return LocalBroadcasterClient{
		broadcasterURL: *u,
	}, nil
}

func (c LocalBroadcasterClient) TranscodeSegment(segment io.Reader, sequenceNumber int64, profiles []video.EncodedProfile, durationMillis int64, manifestID string) (TranscodeResult, error) {
	conf := LivepeerTranscodeConfiguration{
		TimeoutMultiplier: 10,
	}
	conf.Profiles = append(conf.Profiles, profiles...)
	transcodeConfig, err := json.Marshal(&conf)
	if err != nil {
		return TranscodeResult{}, fmt.Errorf("for local B, profiles json encode failed: %v", err)
	}

	return transcodeSegment(segment, sequenceNumber, durationMillis, c.broadcasterURL, manifestID, profiles, string(transcodeConfig))
}
