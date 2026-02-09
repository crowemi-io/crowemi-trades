package config

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"os"

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

type Config struct {
	Alpaca      Alpaca      `json:"alpaca"`
	Crowemi     Crowemi     `json:"crowemi"`
	GoogleCloud GoogleCloud `json:"google_cloud"`
	Firestore   firestore.Client
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
	config.Firestore = *firestoreClient
	// logger
	logger := kitlog.NewLogfmtLogger(kitlog.NewSyncWriter(os.Stderr))
	logger = kitlog.With(logger, "ts", kitlog.DefaultTimestampUTC, "caller", kitlog.DefaultCaller)
	config.Logger = logger

	return &config, nil
}
