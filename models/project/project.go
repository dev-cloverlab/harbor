package project

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/dev-cloverlab/harbor/models/branch"
	"github.com/syndtr/goleveldb/leveldb"
)

type Config struct {
	Name      string           `json:"name"`
	RepoName  string           `json:"repo"`
	Branch    []*branch.Config `json:"branch"`
	CreatedAt time.Time        `json:"created_at"`
}

func (m *Config) Load(name string, db *leveldb.DB) error {
	j, err := db.Get([]byte(name), nil)
	if err != nil {
		return err
	}
	if err := json.Unmarshal(j, m); err != nil {
		return err
	}
	return err
}

func (m *Config) Save(db *leveldb.DB) error {
	j, err := m.ToJson()
	if err != nil {
		return err
	}
	if err := db.Put([]byte(m.Name), j, nil); err != nil {
		return err
	}
	return nil
}

func (m Config) ToJson() ([]byte, error) {
	return json.Marshal(m)
}

func (m Config) KnownBranch(branch string) bool {
	for _, known := range m.Branch {
		if known.Name != branch {
			continue
		}
		return true
	}
	return false
}

func (m Config) GetBranchByName(name string) *branch.Config {
	for _, branch := range m.Branch {
		if branch.Name != name {
			continue
		}
		return branch
	}
	return nil
}

func (m *Config) AddBranch(brName, projName, repoName, harborWorkDir string) {
	m.Branch = append(m.Branch, branch.NewConfig(brName, projName, repoName, harborWorkDir))
}

func (m *Config) DelBranch(name string) {
	for i, branch := range m.Branch {
		if branch.Name != name {
			continue
		}
		tmp := m.Branch[:i]
		m.Branch = append(tmp, m.Branch[i+1:]...)
		break
	}
}

func (m Config) GetGitHubURL(user, token string) string {
	return fmt.Sprintf("https://%s:%s@github.com/%s", user, token, m.RepoName)
}
