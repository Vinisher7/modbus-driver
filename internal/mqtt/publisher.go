package mqtt

import (
	"fmt"
	"modbus-driver/internal/config"
	"modbus-driver/internal/config/logger"
	"time"

	paho "github.com/eclipse/paho.mqtt.golang"
	"go.uber.org/zap"
)

type Publisher struct {
	client paho.Client
}

func NewPublisher(mqtt *config.MQTT) (pub *Publisher, err error) {
	logger.Info("Init NewPublisher", zap.String("journey", "publisher"))

	opts := paho.NewClientOptions().
		AddBroker(mqtt.Broker).
		SetClientID(mqtt.ClientID + "-pub").
		SetCleanSession(true).
		SetAutoReconnect(true).
		SetConnectTimeout(10 * time.Second)

	if mqtt.Username == "" {
		err = fmt.Errorf("mqtt username is empty")
		logger.Error("Getenv func returned an error", err, zap.String("journey", "publisher"))
		return pub, err
	}

	opts.SetUsername(mqtt.Username).SetPassword(mqtt.Password)

	c := paho.NewClient(opts)
	if tok := c.Connect(); tok.Wait() && tok.Error() != nil {
		logger.Error("Connect func returned an error", err, zap.String("journey", "publisher"))
		return pub, fmt.Errorf("error creating a connection with the message broker: %w", tok.Error())
	}

	logger.Info("NewPublisher executed successfully", zap.String("journey", "publisher"))

	return &Publisher{
		client: c,
	}, nil
}

func (p *Publisher) Publish(deviceID, tagID, payload string) {
	logger.Info("Init Publish", zap.String("journey", "publisher"))

	topic := fmt.Sprintf("formatted/erd/%s/%s", deviceID, tagID)

	tok := p.client.Publish(topic, 0, true, payload)

	tok.Wait()

	if err := tok.Error(); err != nil {
		message := fmt.Sprintf("Publish func returned an error on topic=%s", topic)
		logger.Error(message, err, zap.String("journey", "publisher"))
	}

	logger.Info("Publish executed successfully", zap.String("journey", "publisher"))
}
