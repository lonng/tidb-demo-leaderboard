package config

import (
	"strconv"

	"github.com/spf13/pflag"
)

type ServiceOptions struct {
	Port int

	DB struct {
		Host    string
		Port    int
		User    string
		Pass    string
		Name    string
		Options string
	}
	Redis struct {
		Host string
		Pass string
		Port int
	}
}

func (opt *ServiceOptions) AddFlags(flags *pflag.FlagSet) {
	flags.IntVar(&opt.Port, "port", 8080, "Trending service port")

	// Redis server configurations
	flags.StringVar(&opt.Redis.Host, "redis.host", "127.0.0.1", "Redis server host name")
	flags.IntVar(&opt.Redis.Port, "redis.port", 6379, "Redis server port")
	flags.StringVar(&opt.Redis.Pass, "redis.pass", "", "Redis server password")

	// DB server configurations
	flags.StringVar(&opt.DB.Host, "db.host", "127.0.0.1", "Database server host name")
	flags.IntVar(&opt.DB.Port, "db.port", 4000, "Database server port")
	flags.StringVar(&opt.DB.User, "db.user", "root", "Database server user name")
	flags.StringVar(&opt.DB.Pass, "db.pass", "", "Database server password")
	flags.StringVar(&opt.DB.Name, "db.name", "trending", "Database server database name")
	flags.StringVar(&opt.DB.Options, "db.options", "charset=utf8mb4", "Database server connection options")
}

// DSN returns the data source name for the given database.
func (opt *ServiceOptions) DSN() string {
	db := opt.DB
	return db.User + ":" + db.Pass + "@tcp(" + db.Host + ":" + strconv.Itoa(db.Port) + ")/" + db.Name + "?" + db.Options
}
