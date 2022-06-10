package trending

import (
	"database/sql"
	"errors"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/go-redis/redis"
	_ "github.com/go-sql-driver/mysql"
	"github.com/lonng/tidb-demo-trending/config"
	"github.com/pingcap/fn"
)

type Service struct {
	opt *config.ServiceOptions
	rdb *redis.Client
	db  *sql.DB
}

func NewService(opt *config.ServiceOptions) *Service {
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
	s.rdb = rdb

	fmt.Println("Connected to Redis successfully")

	db, err := sql.Open("mysql", s.opt.DSN())
	if err != nil {
		return err
	}
	if err := db.Ping(); err != nil {
		return err
	}
	s.db = db

	fmt.Println("Connected to TiDB successfully")

	// CREATE TABLE
	s.db.Exec(`
		CREATE TABLE IF NOT EXISTS messages(
		    id BIGINT NOT NULL AUTO_RANDOM PRIMARY KEY,
		    text VARCHAR(144),
		    created_at DATETIME DEFAULT NOW(),
		    KEY idx_created_at(created_at)
		);
`)

	fmt.Println("Initialize tables successfully")

	http.Handle("/api/v1/message", fn.Wrap(s.PostMessage))
	http.Handle("/api/v1/messages", fn.Wrap(s.RecentlyMessages))
	http.Handle("/api/v1/top-topics", fn.Wrap(s.TopTopics))

	return http.ListenAndServe(fmt.Sprintf(":%d", s.opt.Port), nil)
}

type (
	MessageRequest struct {
		Text string `json:"text"`
	}

	MessageResponse struct {
		ID        int64     `json:"id"`
		Text      string    `json:"text"`
		CreatedAt time.Time `json:"created_at"`
	}
)

func (s *Service) PostMessage(req *MessageRequest) (*MessageResponse, error) {
	text := strings.TrimSpace(req.Text)
	if text == "" {
		return nil, errors.New("illegal request")
	}
	createdAt := time.Now()
	res, err := s.db.Exec("INSERT INTO messages(text, created_at) VALUES (?, ?)", text, createdAt)
	if err != nil {
		return nil, err
	}
	id, err := res.LastInsertId()
	if err != nil {
		return nil, err
	}

	return &MessageResponse{
		ID:        id,
		Text:      text,
		CreatedAt: createdAt,
	}, nil
}

type (
	RecentlyMessageItem struct {
		ID        uint64 `json:"id"`
		Text      string `json:"text"`
		CreatedAt string `json:"created_at"`
	}
	RecentlyMessageListResponse struct {
		Messages []RecentlyMessageItem `json:"messages"`
	}
)

func (s *Service) RecentlyMessages() (*RecentlyMessageListResponse, error) {
	rows, err := s.db.Query("SELECT id, text, created_at FROM messages ORDER BY created_at DESC LIMIT 100")
	if err != nil {
		return nil, err
	}
	var results []RecentlyMessageItem
	for rows.Next() {
		item := RecentlyMessageItem{}
		err := rows.Scan(&item.ID, &item.Text, &item.CreatedAt)
		if err != nil {
			return nil, err
		}
		results = append(results, item)
	}
	return &RecentlyMessageListResponse{Messages: results}, nil
}

type (
	TopicItem struct {
		Topic string `json:"topic"`
		Hot   int    `json:"hot"`
	}
	TopicListResponse struct {
		Topics []TopicItem `json:"topics"`
	}
)

func (s *Service) TopTopics() (*TopicListResponse, error) {
	cmd := s.rdb.ZRevRangeByScoreWithScores(config.TopicName, redis.ZRangeBy{
		Min:    "-inf",
		Max:    "inf",
		Offset: 0,
		Count:  10,
	})
	res, err := cmd.Result()
	if err != nil {
		return nil, err
	}
	var results []TopicItem
	for _, r := range res {
		results = append(results, TopicItem{
			Topic: r.Member.(string),
			Hot:   int(r.Score),
		})
	}
	return &TopicListResponse{Topics: results}, nil
}
