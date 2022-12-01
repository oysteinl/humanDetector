package cognitiveservices

import (
	"bytes"
	"io/ioutil"
	"os"
	"testing"

	"github.com/joho/godotenv"
)

var endpoint string
var key string

func init() {
	if err := godotenv.Load("../.env"); err != nil {
		panic("No .env file found")
	}

	endpoint, _ = os.LookupEnv("ENDPOINT")
	key, _ = os.LookupEnv("COMPUTERVISION_KEY")
}

func TestCognitiveServicesNoPerson(t *testing.T) {

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
