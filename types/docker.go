package types

type ContainerConfig struct {
	Name         string              `json:"name"`
	Domain       []string            `json:"domain"`
	Image        string              `json:"image"`
	Volumes      map[string]struct{} `json:"volumes"`
	Environments []string            `json:"environments"`
	Port         int64               `json:"port"`
	WorkingDir   string              `json:"workingDir"`
	Cmd          []string            `json:"cmd"`
}
