package branch

import (
	"encoding/json"
	"os"
	"regexp"
	"strings"
	"time"
)

const (
	DeployStateUnknown DeployState = iota
	DeployStateStart
	DeployStateDone
)

type (
	DeployState int

	Config struct {
		Name       string      `json:"name"`
		AllocPorts []int       `json:"port"`
		State      DeployState `json:"state"`
		WorkDir    string      `json:"work"`
		Notice     interface{} `json:"notice"`
		DeployedAt time.Time   `json:"deployed_at"`
	}
)

func NewConfig(brName, projName, repoName, harborWorkDir string) *Config {
	return &Config{
		Name:       brName,
		AllocPorts: []int{},
		State:      DeployStateUnknown,
		WorkDir:    workDirName(brName, projName, repoName, harborWorkDir),
		Notice:     "",
		DeployedAt: time.Time{},
	}
}

func (m Config) GetComposeFilePath() string {
	return strings.Join([]string{m.WorkDir, "docker-compose.yml"}, string(os.PathSeparator))
}

func (m Config) ToJson() ([]byte, error) {
	return json.Marshal(m)
}

func workDirName(brName, projName, repoName, harborWorkDir string) string {
	branchDir := projName + brName
	r := regexp.MustCompile("[^a-z0-9]+")
	branchDir = r.ReplaceAllString(strings.ToLower(branchDir), "")
	return strings.Join([]string{harborWorkDir, repoName, branchDir}, string(os.PathSeparator))
}
