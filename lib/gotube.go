package lib

import (
	"fmt"
	"io"
	"net/http"
	"os"

	"github.com/pkg/errors"
)

var VIDEO_INFO_URL = "https://www.youtube.com/get_video_info?video_id=%s"

// ImportVideo imports the video given
// in the URL inside the dest directory
func ImportVideo(URL, dst string) error {
	vi := &videoInformation{}
	if err := createDestination(dst); err != nil {
		return errors.Wrapf(err, "unable to create destination for: %s", dst)
	}
	if err := vi.getVideoInformation(URL); err != nil {
		return errors.Wrapf(err, "unable to get video information: %s", URL)
	}
	fmt.Printf("downloading %s.%s\n", vi.name, vi.format)
	resp, err := http.Get(vi.URL)
	if err != nil {
		return errors.Wrapf(err, "unable to GET audio: %s", URL)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return errors.New(
			fmt.Sprintf(
				"unable to GET audio: status code different from %d",
				http.StatusOK,
			),
		)
	}

	fh, err := os.Create(fmt.Sprintf("%s/%s.%s", dst, vi.name, vi.format))
	if err != nil {
		return errors.Wrapf(err, "unable to create destination file: %s.mp3", vi.name)
	}
	io.Copy(fh, resp.Body)
	return nil
}

func createDestination(dst string) error {
	if err := os.Mkdir(dst, 0755); err != nil {
		return errors.Wrapf(err, "unable to mkdir %s", dst)
	}
	return nil
}
