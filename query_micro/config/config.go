package config

import "time"

type DatabaseConfig struct {
	Host     string
	Port     int
	User     string
	Password string
	DBName   string
	SSLMode  string
}

type Query struct {
	Name        string `yaml:"name" json:"name"`
	Description string `yaml:"description" json:"description"`
	MinScore    int    `yaml:"minScore" json:"minScore"`
	MaxScore    int    `yaml:"maxScore" json:"maxScore"`
}

type Answer struct {
	UserId    int       `yaml:"userId" json:"userId"`
	QueryName string    `yaml:"queryName" json:"queryName"`
	Login     string    `yaml:"login" json:"login"`
	Score     int       `yaml:"score" json:"score"`
	Answer    string    `yaml:"answer" json:"answer"`
	CreatedAt time.Time `yaml:"createdAt" json:"createdAt"`
}

type QueryForUser struct {
	Name        string `yaml:"name" json:"name"`
	Description string `yaml:"description" json:"description"`
	MinScore    int    `yaml:"minScore" json:"minScore"`
	MaxScore    int    `yaml:"maxScore" json:"maxScore"`
	Score       int    `yaml:"score" json:"score"`
	Answer      string `yaml:"answer" json:"answer"`
}

type UsersForQuery struct {
	Name        string `yaml:"name" json:"name"`
	Description string `yaml:"description" json:"description"`
	MinScore    int    `yaml:"minScore" json:"minScore"`
	MaxScore    int    `yaml:"maxScore" json:"maxScore"`
	Login       string `yaml:"login" json:"login"`
	Answer      string `yaml:"answer" json:"answer"`
	Score       int    `yaml:"score" json:"score"`
}

type QueryStats struct {
	TotalAnswers int64
	AverageScore float64
	MinScore     int
	MaxScore     int
}

type AnswersForQuery struct {
	Name        string `yaml:"name" json:"name"`
	Description string `yaml:"description" json:"description"`
	MinScore    int    `yaml:"minScore" json:"minScore"`
	MaxScore    int    `yaml:"maxScore" json:"maxScore"`
	Login       string `yaml:"login" json:"login"`
	Answer      string `yaml:"answer" json:"answer"`
	Score       int    `yaml:"score" json:"score"`
	UserId      int    `yaml:"userId" json:"userId"`
}
