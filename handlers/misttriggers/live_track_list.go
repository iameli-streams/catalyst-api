package misttriggers

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/golang/glog"
	"github.com/livepeer/catalyst-api/clients"
	"github.com/livepeer/catalyst-api/errors"
)

type LiveTrackListPayload struct {
	StreamName string
	TrackList  map[string]clients.MistStreamInfoTrack
}

func (payload *LiveTrackListPayload) CountVideoTracks() int {
	res := 0
	for _, td := range payload.TrackList {
		if td.Type == "video" {
			res++
		}
	}
	return res
}

func ParseLiveTrackListPayload(payload MistTriggerBody) (LiveTrackListPayload, error) {
	lines := payload.Lines()
	if len(lines) != 2 {
		return LiveTrackListPayload{}, fmt.Errorf("expected 2 lines in LIVE_TRACK_LIST payload but got lines=%d payload=%s", len(lines), payload)
	}

	tl := map[string]clients.MistStreamInfoTrack{}
	err := json.Unmarshal([]byte(lines[1]), &tl)
	if err != nil {
		return LiveTrackListPayload{}, fmt.Errorf("error unmarhsaling LIVE_TRACK_LIST payload err=%s payload=%s", err, payload)
	}

	return LiveTrackListPayload{
		StreamName: lines[0],
		TrackList:  tl,
	}, nil
}

func (d *MistCallbackHandlersCollection) TriggerLiveTrackList(ctx context.Context, w http.ResponseWriter, req *http.Request, payload MistTriggerBody) {
	body, err := ParseLiveTrackListPayload(payload)
	if err != nil {
		glog.Infof("Error parsing LIVE_TRACK_LIST payload error=%q payload=%q", err, string(payload))
		errors.WriteHTTPBadRequest(w, "Error parsing LIVE_TRACK_LIST payload", err)
		return
	}
	err = d.broker.TriggerLiveTrackList(ctx, &body)
	if err != nil {
		glog.Infof("Error handling LIVE_TRACK_LIST payload error=%q payload=%q", err, string(payload))
		errors.WriteHTTPInternalServerError(w, "Error handling LIVE_TRACK_LIST payload", err)
		return
	}
}
