package leaderboard

import (
	"bytes"
	"database/sql"
	_ "embed"
	"errors"
	"fmt"
	"math/rand"
	"net/http"
	"strings"
	"time"

	"github.com/go-redis/redis"
	_ "github.com/go-sql-driver/mysql"
	"github.com/gorilla/mux"
	"github.com/lonng/tidb-demo-leaderboard/config"
	"github.com/pingcap/fn"
)

//go:embed app.html
var appHtml []byte

var symbols = []string{
	"🏠", "🏡", "🏫", "🏢", "🏣", "🏥", "🏦", "🏪", "🏩", "🏨", // "💒", "⛪️", "🏬", "🏤", "🌇", "🕌", "🕍", "🌆", "🏯", "🏰", "⛺️", "🏭", "🗼", "🏘", "🏞", "🏟", "🏙", "🏚", "🏛", "🏗", "🛖", "🪦",
}

var symbolScores = map[string]int{} // symbol => score

func init() {
	for _, s := range symbols {
		symbolScores[s] = rand.Intn(100)
	}
}

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
	_, err = s.db.Exec(`
		CREATE TABLE IF NOT EXISTS user_score(
		    id BIGINT NOT NULL AUTO_RANDOM PRIMARY KEY,
		    name VARCHAR(32),
		    score BIGINT DEFAULT 0,
		    created_at DATETIME DEFAULT NOW(),
		    UNIQUE KEY idx_name(name)
		);

		
`)
	if err != nil {
		return err
	}
	_, err = s.db.Exec(`
		CREATE TABLE IF NOT EXISTS score_histories(
		    id BIGINT NOT NULL AUTO_RANDOM PRIMARY KEY,
		    name VARCHAR(32),
		    score_changed BIGINT,
		    created_at DATETIME DEFAULT NOW()
		);`)
	if err != nil {
		return err
	}

	fmt.Println("Initialize tables successfully")

	startTime := time.Now()

	router := mux.NewRouter()
	router.Handle("/api/v1/join", fn.Wrap(s.Join)).Methods(http.MethodPost)
	router.Handle("/api/v1/round", fn.Wrap(s.NextRound)).Methods(http.MethodGet)
	router.Handle("/api/v1/round", fn.Wrap(s.RoundResult)).Methods(http.MethodPost)
	router.Handle("/api/v1/leaderboard", fn.Wrap(s.Leaderboard))
	router.Handle("/", http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		http.ServeContent(writer, request, "app.html", startTime, bytes.NewReader(appHtml))
	}))

	return http.ListenAndServe(fmt.Sprintf(":%d", s.opt.Port), router)
}

type RoundResponse struct {
	LeftSymbol  string `json:"left_symbol"`
	RightSymbol string `json:"right_symbol"`
}

func randomSymbol() *RoundResponse {
	n1 := rand.Intn(len(symbols))
	n2 := rand.Intn(len(symbols))
	for n1 == n2 {
		n2 = rand.Intn(len(symbols))
	}

	return &RoundResponse{
		LeftSymbol:  symbols[n1],
		RightSymbol: symbols[n2],
	}
}

func (s *Service) NextRound() (*RoundResponse, error) {
	return randomSymbol(), nil
}

type (
	JoinRequest struct {
		Name string `json:"name"`
	}

	JoinResponse struct {
		Name  string `json:"name"`
		Score int64  `json:"score"`
	}
)

func (s *Service) Join(req *JoinRequest) (*JoinResponse, error) {
	name := strings.TrimSpace(req.Name)
	rows, err := s.db.Query("SELECT score FROM user_score WHERE name = ?", name)
	if err != nil || !rows.Next() {
		return &JoinResponse{Name: name}, nil
	}
	defer rows.Close()

	var score int64
	err = rows.Scan(&score)
	if err != nil {
		return nil, err
	}
	return &JoinResponse{
		Name:  name,
		Score: score,
	}, nil
}

type (
	RoundResultRequest struct {
		Name        string `json:"name"`
		LeftSymbol  string `json:"left_symbol"`
		RightSymbol string `json:"right_symbol"`
		ChooseLeft  bool   `json:"choose_left"`
	}

	RoundResultResponse struct {
		IsWin        bool           `json:"is_win"`
		ChangedScore int64          `json:"changed_score"`
		NextRound    *RoundResponse `json:"next_round"`
	}
)

func (s *Service) RoundResult(req *RoundResultRequest) (*RoundResultResponse, error) {
	leftScore, found := symbolScores[req.LeftSymbol]
	if !found {
		return nil, errors.New("illegal request")
	}
	rightScore, found := symbolScores[req.RightSymbol]
	if !found {
		return nil, errors.New("illegal request")
	}
	isWin := (req.ChooseLeft && leftScore > rightScore) || (!req.ChooseLeft && leftScore < rightScore)
	changedScore := int64(-1)
	if isWin {
		changedScore = 1
	}

	tx, err := s.db.Begin()
	if err != nil {
		return nil, err
	}
	name := strings.TrimSpace(req.Name)
	_, err = tx.Exec("INSERT INTO user_score (name, score) VALUES (?, ?) ON DUPLICATE KEY UPDATE score = score + ?",
		name, changedScore, changedScore)
	if err != nil {
		return nil, tx.Rollback()
	}

	_, err = tx.Exec("INSERT INTO score_histories(name, score_changed) VALUES (?, ?)", name, changedScore)
	if err != nil {
		return nil, tx.Rollback()
	}

	return &RoundResultResponse{
		IsWin:        isWin,
		ChangedScore: changedScore,
		NextRound:    randomSymbol(),
	}, tx.Commit()
}

type (
	ScoreItem struct {
		Name  string `json:"name"`
		Score int    `json:"score"`
	}
	ScoreRankResponse struct {
		Rank []ScoreItem `json:"rank"`
	}
)

func (s *Service) Leaderboard() (*ScoreRankResponse, error) {
	cmd := s.rdb.ZRevRangeByScoreWithScores(config.Leaderboard, redis.ZRangeBy{
		Min:    "-inf",
		Max:    "inf",
		Offset: 0,
		Count:  10,
	})
	res, err := cmd.Result()
	if err != nil {
		return nil, err
	}
	var results []ScoreItem
	for _, r := range res {
		results = append(results, ScoreItem{
			Name:  r.Member.(string),
			Score: int(r.Score),
		})
	}
	return &ScoreRankResponse{Rank: results}, nil
}
