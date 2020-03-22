package lib

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"strings"

	"github.com/pkg/errors"
)

// PlayerResponse is the the data
// with the YT player information
type PlayerResponse struct {
	StreamingData Data    `json:"streamingData"`
	VideoDetails  Details `json:"videoDetails"`
}

// Details is the details of the video
// title, authors, etc.
type Details struct {
	Title string `json:"title"`
}

// Data is an array of Format
type Data struct {
	Formats []Format `json:"adaptiveFormats"`
}

// Format is the different set of data
// composing a video (codec, video, audio, etc.)
type Format struct {
	URL      string `json:"url"`
	MimeType string `json:"mimeType"`
}

type videoInformation struct {
	name   string
	URL    string
	ID     string
	format string
}

func (vi *videoInformation) getVideoID(URL string) error {
	if vi.ID != "" {
		return nil
	}
	u, err := url.Parse(URL)
	if err != nil {
		return errors.Wrapf(err, "unable to parse URL: %s", URL)
	}
	q := u.Query()
	if v, ok := q["v"]; ok {
		vi.ID = v[0]
		return nil
	}
	return errors.New("missing video ID in the URL")
}

func (vi *videoInformation) getVideoInformation(URL string) error {
	if err := vi.getVideoID(URL); err != nil {
		return errors.Wrapf(err, "unable to get video ID from %s", URL)
	}
	resp, err := http.Get(fmt.Sprintf(VIDEO_INFO_URL, vi.ID))
	if err != nil {
		return errors.Wrapf(err, "unable to GET video information: %s", URL)
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return errors.Wrapf(err, "unable to extract video information: %s", URL)
	}

	q, err := url.ParseQuery(string(body))
	if err != nil {
		return errors.Wrapf(err, "unable to parse video information: %s", URL)
	}

	var player PlayerResponse
	playerResponse, ok := q["player_response"]
	if !ok {
		return errors.New("player response is not available")
	}
	if err := json.Unmarshal([]byte(playerResponse[0]), &player); err != nil {
		return errors.Wrap(err, "unable to extract player response")
	}
	vi.name = player.VideoDetails.Title

	for _, f := range player.StreamingData.Formats {
		if strings.Contains(f.MimeType, "audio/mp4") {
			vi.URL = f.URL
			vi.format = "mp4"
			return nil
		}
	}
	return nil
}
