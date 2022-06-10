package config

import "github.com/spf13/pflag"

type ConsumerOptions struct {
	Kafka struct {
		Server string
		Topic  string
	}
	Redis struct {
		Host string
		Port int
		Pass string
	}
}

func (opt *ConsumerOptions) AddFlags(flags *pflag.FlagSet) {
	flags.StringVar(&opt.Kafka.Server, "kafka.server", "127.0.0.1:9092", "Kafka bootstrap server address")
	flags.StringVar(&opt.Kafka.Topic, "kafka.topic", "trending", "Kafka topic name")

	flags.StringVar(&opt.Redis.Host, "redis.host", "127.0.0.1", "Redis server host name")
	flags.IntVar(&opt.Redis.Port, "redis.port", 6379, "Redis server port")
	flags.StringVar(&opt.Redis.Pass, "redis.pass", "", "Redis server password")
}
