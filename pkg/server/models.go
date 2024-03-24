// Package server implements the methods and data structures
// responsible for the operation of the HTTP server
// All data structures are defined in a separate models.go file
package server

import (
	"context"
	"github.com/redis/go-redis/v9"
	"log"
	"net/http"
	"time"
)

type S struct {
	S                         *http.Server
	c                         Config
	db                        *redis.Client
	LogInfo, LogErr, LogFatal *log.Logger
}

type FloodControl interface {
	Check(ctx context.Context, userID int64) (bool, error)
}

// Config is structure for reading the configuration yaml file
type Config struct {
	Server struct {
		Host string `yaml:"host"`
		Port string `yaml:"port"`

		FloodControl struct {
			Time     time.Duration `yaml:"time"`
			Requests int           `yaml:"requests"`
		} `yaml:"flood_control"`
	} `yaml:"server"`
	DB struct {
		Host string `yaml:"host"`
		Port string `yaml:"port"`
	} `yaml:"redis"`
}
