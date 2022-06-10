package consumer

import (
	"fmt"

	"github.com/confluentinc/confluent-kafka-go/kafka"
	"github.com/go-redis/redis"
	"github.com/lonng/tidb-demo-trending/config"
)

type Service struct {
	opt *config.ConsumerOptions
}

func NewService(opt *config.ConsumerOptions) *Service {
	return &Service{
		opt: opt,
	}
}

func (s *Service) Serve() error {
	rdb := redis.NewClient(&redis.Options{
		Addr:     fmt.Sprintf("%s:%d", s.opt.Redis.Host, s.opt.Redis.Port),
		Password: s.opt.Redis.Pass,
		DB:       0, // use default DB
	})
	result := rdb.Ping()
	if err := result.Err(); err != nil {
		return err
	}
	defer rdb.Close()

	fmt.Println("Connected to Redis successfully")

	c, err := kafka.NewConsumer(&kafka.ConfigMap{
		"bootstrap.servers": s.opt.Kafka.Server,
		"group.id":          "message-consumer",
		"auto.offset.reset": "earliest",
	})
	if err != nil {
		return err
	}
	err = c.SubscribeTopics([]string{s.opt.Kafka.Topic}, nil)
	if err != nil {
		return err
	}
	defer c.Close()

	for {
		msg, err := c.ReadMessage(-1)
		if err == nil {
			fmt.Printf("Message on %s: %s\n", msg.TopicPartition, string(msg.Value))
		} else {
			// The client will automatically try to recover from all errors.
			fmt.Printf("Consumer error: %v (%v)\n", err, msg)
		}
	}
}
