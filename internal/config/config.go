package config

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"os"
	"strings"

	kitlog "github.com/go-kit/log"

	"cloud.google.com/go/firestore"
)

type GoogleCloudFirestore struct {
	Database string `json:"database" omitempty:"true"`
}

type GoogleCloud struct {
	OrganizationID string               `json:"organization_id" omitempty:"true"`
	ProjectID      string               `json:"project_id" omitempty:"true"`
	Region         string               `json:"region" omitempty:"true"`
	Firestore      GoogleCloudFirestore `json:"firestore" omitempty:"true"`
}

type Alpaca struct {
	APIKey       string `json:"api_key" omitempty:"true"`
	APISecretKey string `json:"api_secret_key" omitempty:"true"`
	APIBaseURL   string `json:"api_base_url" omitempty:"true"`
	APIDataURL   string `json:"api_data_url" omitempty:"true"`
}

type Crowemi struct {
	ClientName      string            `json:"client_name" omitempty:"true"`
	ClientID        string            `json:"client_id" omitempty:"true"`
	ClientSecretKey string            `json:"client_secret_key" omitempty:"true"`
	Uri             map[string]string `json:"uri" omitempty:"true"`
	DatabaseURI     string            `json:"database_uri" omitempty:"true"`
}

type Scheduler struct {
	ScheduleTasks []SchedulerTask `json:"scheduler_tasks" omitempty:"true"`
}

type SchedulerTask struct {
	Name     string `json:"name" omitempty:"true"`
	Schedule string `json:"schedule" omitempty:"true"`
}

func (s Scheduler) ScheduleForTask(taskName string) string {
	for _, task := range s.ScheduleTasks {
		if strings.EqualFold(strings.TrimSpace(task.Name), strings.TrimSpace(taskName)) {
			return strings.TrimSpace(task.Schedule)
		}
	}
	return ""
}

type Runtime struct {
	PortfolioID        string    `json:"portfolio_id" omitempty:"true"`
	HTTPListenAddr     string    `json:"http_listen_addr" omitempty:"true"`
	Scheduler          Scheduler `json:"scheduler" omitempty:"true"`
	TaskTimeout        string    `json:"task_timeout" omitempty:"true"`
	StreamReconnectMin string    `json:"stream_reconnect_min" omitempty:"true"`
	StreamReconnectMax string    `json:"stream_reconnect_max" omitempty:"true"`
}

type Notifier struct {
	Telegram *Telegram `json:"telegram" omitempty:"true"`
}

type Telegram struct {
	BotToken string `json:"bot_token" omitempty:"true"`
	ChatID   int64  `json:"chat_id" omitempty:"true"`
}

type Config struct {
	Alpaca      Alpaca      `json:"alpaca"`
	Crowemi     Crowemi     `json:"crowemi"`
	GoogleCloud GoogleCloud `json:"google_cloud"`
	Notifier    Notifier    `json:"notifier" omitempty:"true"`
	Runtime     Runtime     `json:"runtime"`
	Firestore   *firestore.Client
	Logger      kitlog.Logger
}

func Bootstrap(configPath string) (*Config, error) {
	var config Config
	value := os.Getenv("CONFIG")
	if value != "" {
		decode, err := base64.StdEncoding.DecodeString(value)
		if err != nil {
			return nil, err
		}
		json.Unmarshal(decode, &config)
	} else {
		contents, err := os.ReadFile(configPath)
		if err != nil {
			return nil, err
		}
		json.Unmarshal(contents, &config)
	}
	// firestore
	firestoreClient, err := firestore.NewClientWithDatabase(context.Background(), config.GoogleCloud.ProjectID, config.GoogleCloud.Firestore.Database)
	if err != nil {
		return nil, err
	}
	config.Firestore = firestoreClient
	// logger
	logger := kitlog.NewLogfmtLogger(kitlog.NewSyncWriter(os.Stderr))
	logger = kitlog.With(logger, "ts", kitlog.DefaultTimestampUTC, "caller", kitlog.DefaultCaller)
	config.Logger = logger
	applyRuntimeDefaults(&config)

	return &config, nil
}

func applyRuntimeDefaults(config *Config) {
	if config.Runtime.HTTPListenAddr == "" {
		config.Runtime.HTTPListenAddr = ":8080"
	}
	if config.Runtime.TaskTimeout == "" {
		config.Runtime.TaskTimeout = "30s"
	}
	if config.Runtime.StreamReconnectMin == "" {
		config.Runtime.StreamReconnectMin = "1s"
	}
	if config.Runtime.StreamReconnectMax == "" {
		config.Runtime.StreamReconnectMax = "30s"
	}
}
