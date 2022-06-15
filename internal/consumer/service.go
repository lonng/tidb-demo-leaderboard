package consumer

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"strconv"

	"github.com/go-redis/redis"
	"github.com/lonng/tidb-demo-leaderboard/config"
	"github.com/pingcap/errors"
	"github.com/segmentio/kafka-go"
)

type Service struct {
	opt *config.ConsumerOptions
	rdb *redis.Client
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
		return errors.Annotate(err, "ping redis failed")
	}
	s.rdb = rdb
	defer rdb.Close()

	fmt.Println("Connected to Redis successfully")

	conn, err := kafka.DialLeader(context.Background(), "tcp", s.opt.Kafka.Server, s.opt.Kafka.Topic, 0)
	if err != nil {
		log.Fatal("failed to dial leader:", err)
	}
	defer conn.Close()

	fmt.Println("Connected to Kafka successfully")

	for {
		message, err := conn.ReadMessage(10e3)
		if err != nil {
			fmt.Printf("Consumer error: %v\n", err)
			continue
		}
		fmt.Println("Message on", string(message.Value))
		s.updateCache(message.Value)
	}
}

type Message struct {
	Data []struct {
		Name  string `json:"name"`
		Score string `json:"score"`
	} `json:"data"`
}

func (s *Service) updateCache(val []byte) {
	m := Message{}
	if err := json.Unmarshal(val, &m); err != nil {
		fmt.Println("Unmarshal message failed", err)
		return
	}

	if len(m.Data) == 0 {
		return
	}

	for _, item := range m.Data {
		score, err := strconv.ParseInt(item.Score, 10, 64)
		if err != nil {
			continue
		}
		cmd := s.rdb.ZAdd(config.Leaderboard, redis.Z{Score: float64(score), Member: item.Name})
		if err := cmd.Err(); err != nil {
			fmt.Println("ZIncrBy failed", item.Name, item.Score, err)
		}
	}
}
