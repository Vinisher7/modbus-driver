package mqtt

import (
	"fmt"
	"log"
	"time"

	"modbus-driver/internal/config"

	paho "github.com/eclipse/paho.mqtt.golang"
)

type Publisher struct {
	client paho.Client
}

func NewPublisher(cfg *config.Config) (*Publisher, error) {
	opts := paho.NewClientOptions().
		AddBroker(cfg.MQTTBroker).
		SetClientID(cfg.MQTTClientID).
		SetCleanSession(true).
		SetAutoReconnect(true).
		SetConnectTimeout(10 * time.Second)

	if cfg.MQTTUsername != "" {
		opts.SetUsername(cfg.MQTTUsername).SetPassword(cfg.MQTTPassword)
	}

	c := paho.NewClient(opts)
	if tok := c.Connect(); tok.Wait() && tok.Error() != nil {
		return nil, fmt.Errorf("mqtt connect: %w", tok.Error())
	}
	log.Println("[MQTT] connected to", cfg.MQTTBroker)
	return &Publisher{client: c}, nil
}

// Publish envia payload JSON para o tópico correto
// tópico: /devices/formatted_data/erd/{deviceUUID}/{tagName}
func (p *Publisher) Publish(deviceUUID, tagID, payload string) {
	topic := fmt.Sprintf("/devices/formatted_data/erd/%s/%s", deviceUUID, tagID)
	tok := p.client.Publish(topic, 0, true, payload)
	tok.Wait()
	if err := tok.Error(); err != nil {
		log.Printf("[MQTT] publish error topic=%s err=%v", topic, err)
	}
}
