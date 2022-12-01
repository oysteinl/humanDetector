package camera

import (
	"io"
	"net/http"
)

func FetchSnapshot(snapUrl string) (io.ReadCloser, error) {

	resp, err := http.Get(snapUrl)
	if err != nil {
		return nil, err
	}
	return resp.Body, nil
}
