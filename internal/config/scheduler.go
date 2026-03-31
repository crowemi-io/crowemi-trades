package config

type Scheduler struct {
	Enabled bool   `json:"enabled" omitempty:"true"`
	Tasks   []Task `json:"tasks" omitempty:"true"`
}

type TaskOptions struct {
	WriteFile bool `json:"writeFile" omitempty:"true"`
}

type Task struct {
	Name     string      `json:"name" omitempty:"true"`
	Schedule string      `json:"schedule" omitempty:"true"`
	Enabled  bool        `json:"enabled" omitempty:"true"`
	Options  TaskOptions `json:"options" omitempty:"true"`
}
