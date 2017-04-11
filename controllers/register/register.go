package register

import (
	"encoding/json"
	"fmt"
	"net/http"

	myctx "github.com/dev-cloverlab/harbor/middleware/context"

	"github.com/dev-cloverlab/harbor/models/branch"
	"github.com/dev-cloverlab/harbor/models/project"
	"github.com/syndtr/goleveldb/leveldb"
)

type Request struct {
	Name     string `json:"name"`
	RepoName string `json:"repo"`
}

func (m Request) IsValid() error {
	if m.Name == "" {
		return fmt.Errorf("name is required")
	}
	if m.RepoName == "" {
		return fmt.Errorf("repo is required")
	}
	return nil
}

func Handler(db *leveldb.DB, w http.ResponseWriter, r *http.Request) {

	// load global settings from request context
	globalConf := &myctx.Config{}
	globalConf.FromJson(r.Context().Value(myctx.ConfigKey).([]byte))

	// request validation
	payload := r.PostFormValue("payload")

	req := Request{}
	if err := json.Unmarshal([]byte(payload), &req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	if err := req.IsValid(); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	projConf := project.Config{
		Name:      req.Name,
		RepoName:  req.RepoName,
		Branch:    []*branch.Config{},
		CreatedAt: globalConf.Timestamp,
	}
	projConf.Save(db)

	j, err := projConf.ToJson()
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	w.Write(j)
}
