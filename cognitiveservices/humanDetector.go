package cognitiveservices

import (
	"context"
	"github.com/Azure/azure-sdk-for-go/services/cognitiveservices/v2.1/computervision"
	"github.com/Azure/go-autorest/autorest"
	log "github.com/sirupsen/logrus"
	"io"
)

var computerVisionContext context.Context

func PersonIsDetected(endpoint string, key string, imageStream io.ReadCloser) (bool, error) {

	computerVisionClient := computervision.New(endpoint)
	computerVisionClient.Authorizer = autorest.NewCognitiveServicesAuthorizer(key)
	computerVisionContext = context.Background()
	imageAnalysis, err := computerVisionClient.DetectObjectsInStream(computerVisionContext, imageStream)
	if err != nil {
		return false, err
	}
	var objects []string
	var hit = false
	for _, object := range *imageAnalysis.Objects {
		objects = append(objects, *object.Object)
		if *object.Object == "person" || *object.Object == "animal" {
			hit = true
		}
	}
	log.Debug(objects)
	return hit, nil
}
