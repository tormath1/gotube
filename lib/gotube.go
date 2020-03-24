package lib

import (
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"net/url"
	"os"
	"regexp"

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

// ImportPlaylist fetchs all the video
// in a playlist then download them
func ImportPlaylist(URL, dst string) error {
	pi := &playlistInformation{}
	if err := pi.getPlaylistID(URL); err != nil {
		return errors.Wrapf(err, "unable to get playlist ID %s", URL)
	}
	if err := pi.getPlaylistVideoURLs(URL); err != nil {
		return errors.Wrapf(err, "unable to get playlist video URLs %s", URL)
	}
	for _, u := range pi.URLs {
		if err := ImportVideo(u, dst); err != nil {
			fmt.Printf("unable to download video %s: %v", u, err)
		}
	}
	return nil
}

type playlistInformation struct {
	URLs []string
	ID   string
	Name string
}

func (pi *playlistInformation) getPlaylistID(URL string) error {
	if pi.ID != "" {
		return nil
	}
	u, err := url.Parse(URL)
	if err != nil {
		return errors.Wrapf(err, "unable to parse URL: %s", URL)
	}
	q := u.Query()
	if v, ok := q["list"]; ok {
		pi.ID = v[0]
		return nil
	}
	return errors.New("missing playlist ID in the URL")
}

func (pi *playlistInformation) getPlaylistVideoURLs(URL string) error {
	resp, err := http.Get(URL)
	if err != nil {
		return errors.Wrapf(err, "unable to GET playlist information: %s", URL)
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return errors.Wrapf(err, "unable to extract video information: %s", URL)
	}

	re := regexp.MustCompile(fmt.Sprintf(`watch\?v=\S+?list=%s`, pi.ID))
	URLs := re.FindAllString(string(body), -1)
	pi.URLs = removeDuplicate(URLs)
	return nil
}

func removeDuplicate(URLs []string) []string {
	d := make(map[string]bool, 0)
	res := make([]string, 0)
	for _, URL := range URLs {
		if _, ok := d[URL]; !ok {
			d[URL] = true
			res = append(res, URL)
		}
	}
	return res

}

func createDestination(dst string) error {
	if _, err := os.Stat(dst); os.IsNotExist(err) {
		if err := os.Mkdir(dst, 0755); err != nil {
			return errors.Wrapf(err, "unable to mkdir %s", dst)
		}
	}
	return nil
}
