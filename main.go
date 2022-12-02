package main

import (
	"bytes"
	"fmt"
	"humanDetector/camera"
	"humanDetector/cognitiveservices"
	"io"
	"io/ioutil"
	"os"
	"sync"
	"time"

	mqtt "github.com/eclipse/paho.mqtt.golang"
	"github.com/joho/godotenv"
	notification "github.com/oysteinl/notificationManager"
	"github.com/oysteinl/notificationManager/notifier"
	log "github.com/sirupsen/logrus"
)

var endpoint string
var endpointExists bool
var key string
var keyExists bool
var mqttUrl string
var mqttUrlExists bool
var mqttMotionTopic string
var mqttMotionTopicExists bool
var mqttAwayTopic string
var mqttAwayTopicExists bool
var mqttUser string
var mqttUserExists bool
var mqttPass string
var mqttPassExists bool
var telegramChatId string
var telegramChatIdExists bool
var telegramToken string
var telegramTokenExists bool
var snapUrl string
var snapUrlExists bool
var failFolder string
var failFolderExists bool

const AWAY = "not_home"
const MOTION = "ON"

var safeAwayTracker SafeAwayTracker

type SafeAwayTracker struct {
	mu   sync.Mutex
	away bool
}

func (c *SafeAwayTracker) SetAway(away bool) {
	c.mu.Lock()
	c.away = away
	c.mu.Unlock()
}

func init() {
	log.SetLevel(log.DebugLevel)
	log.SetFormatter(&log.TextFormatter{
		DisableColors: false,
		FullTimestamp: true,
	})
	safeAwayTracker = SafeAwayTracker{}

	envPath := "env_example"
	if len(os.Args) == 2 {
		envPath = os.Args[1]
	}
	if err := godotenv.Load(envPath); err != nil {
		panic("No .env file found")
	}
	endpoint, endpointExists = os.LookupEnv("ENDPOINT")
	key, keyExists = os.LookupEnv("COMPUTERVISION_KEY")
	mqttUrl, mqttUrlExists = os.LookupEnv("MQTT_URL")
	mqttMotionTopic, mqttMotionTopicExists = os.LookupEnv("MQTT_MOTION_TOPIC")
	mqttAwayTopic, mqttAwayTopicExists = os.LookupEnv("MQTT_AWAY_TOPIC")
	mqttUser, mqttUserExists = os.LookupEnv("MQTT_USER")
	mqttPass, mqttPassExists = os.LookupEnv("MQTT_PASSWORD")
	telegramChatId, telegramChatIdExists = os.LookupEnv("TELEGRAM_CHATID")
	telegramToken, telegramTokenExists = os.LookupEnv("TELEGRAM_TOKEN")
	snapUrl, snapUrlExists = os.LookupEnv("SNAP_URL")
	failFolder, failFolderExists = os.LookupEnv("FAIL_FOLDER")

	if !endpointExists || !keyExists || !mqttUrlExists || !mqttMotionTopicExists || !mqttAwayTopicExists || !mqttUserExists || !mqttPassExists || !telegramChatIdExists || !telegramTokenExists || !snapUrlExists || !failFolderExists {
		panic("Env variables not set")
	}
}

func main() {

	log.Info("Starting")
	keepAlive := make(chan os.Signal)
	listen()
	<-keepAlive
}

func listen() {
	client := connectMQTT("humanDetector")
	client.Subscribe(mqttMotionTopic, 0, mqttMotionCallback)
	client.Subscribe(mqttAwayTopic, 0, mqttAwayCallback)
}

func mqttMotionCallback(client mqtt.Client, msg mqtt.Message) {
	payload := string(msg.Payload())
	if thereIsMotion(payload) && safeAwayTracker.away {
		log.Info("Motion detected and we are not home")
		buf, isDetected, err := personOutside(3)
		if err != nil {
			log.Error(err)
			return
		}
		if isDetected {
			log.Info("Person detected")
			data, _ := ioutil.ReadAll(buf)
			err = notification.NotifyTelegram(notifier.SendPhoto, "Person detected", data, notifier.TelegramConfig{ChatId: telegramChatId, Token: telegramToken})
			if err != nil {
				log.Error(err)
			}
		} else {
			log.Info("Person not detected")
			error := ioutil.WriteFile(failFolder+time.Now().Format("20060102T150405")+".jpeg", buf.Bytes(), 0644)
			if error != nil {
				log.Error(error)
			}
		}
	}
}

func personOutside(retries int) (*bytes.Buffer, bool, error) {

	buf := &bytes.Buffer{}
	for i := 1; i <= retries; i++ {
		log.Infof("Take %d", i)
		imageStream, err := camera.FetchSnapshot(snapUrl)
		if err != nil {
			log.Error("Error fetching snap from camera", err)
			return nil, false, err
		}
		tee := io.TeeReader(imageStream, buf)
		isDetected, err := cognitiveservices.PersonIsDetected(endpoint, key, io.NopCloser(tee))
		if err != nil {
			log.Error("error from cognitive services", err)
			return nil, false, err
		}
		if isDetected {
			return buf, isDetected, nil
		}
		time.Sleep(2 * time.Second)
	}
	return buf, false, nil
}

func thereIsMotion(payload string) bool {
	return payload == MOTION
}

func mqttAwayCallback(client mqtt.Client, msg mqtt.Message) {
	payload := string(msg.Payload())
	log.Debugf("* [%s] %s\n", msg.Topic(), payload)
	safeAwayTracker.SetAway(payload == AWAY)
}

func connectMQTT(clientId string) mqtt.Client {
	opts := createClientOptions(clientId)
	client := mqtt.NewClient(opts)
	token := client.Connect()
	for !token.WaitTimeout(3 * time.Second) {
	}
	if err := token.Error(); err != nil {
		log.Fatal(err)
	}
	return client
}

func createClientOptions(clientId string) *mqtt.ClientOptions {
	opts := mqtt.NewClientOptions()
	opts.AddBroker(fmt.Sprintf("tcp://%s:%d", mqttUrl, 1883))
	opts.SetUsername(mqttUser)
	opts.SetPassword(mqttPass)
	opts.SetClientID(clientId)
	return opts
}
