package cognitiveservices

import (
	"bytes"
	"io/ioutil"
	"testing"
)

var endpoint string
var key string

func init() {

	endpoint = ""
	key = ""
}

func TestCognitiveServicesNoPerson(t *testing.T) {

	if len(endpoint) == 0 || len(key) == 0 {
		return
	}

	data, err := ioutil.ReadFile("../testdata/snap.jpeg")
	if err != nil {
		t.Fatal(err)
	}

	reader := bytes.NewReader(data)
	readCloser := ioutil.NopCloser(reader)
	defer readCloser.Close()
	isDetected, err := PersonIsDetected(endpoint, key, readCloser)
	if err != nil {
		t.Fatal(err)
	}
	if isDetected {
		t.Fatal("should not detect any person")
	}
}
