package config

type Watcher struct {
	Enabled bool `json:"enabled" omitempty:"true"`
}

type Updater struct {
	Enabled bool `json:"enabled" omitempty:"true"`
}

type Streamer struct {
	Watcher Watcher `json:"watcher" omitempty:"true"`
	Updater Updater `json:"updater" omitempty:"true"`
}
