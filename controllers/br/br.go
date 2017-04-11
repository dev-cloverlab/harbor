package br

import (
	"net/http"

	myctx "github.com/dev-cloverlab/harbor/middleware/context"

	"github.com/dev-cloverlab/harbor/models/branch"
	"github.com/dev-cloverlab/harbor/models/project"
	"github.com/syndtr/goleveldb/leveldb"
)

func Handler(db *leveldb.DB, w http.ResponseWriter, r *http.Request) {

	name := r.URL.Query().Get("name")
	branchName := r.URL.Query().Get("branch")

	if name == "" {
		http.Error(w, "name is required", http.StatusBadRequest)
		return
	}

	if branchName == "" {
		http.Error(w, "branch is required", http.StatusBadRequest)
		return
	}

	// load global settings from request context
	globalConf := &myctx.Config{}
	globalConf.FromJson(r.Context().Value(myctx.ConfigKey).([]byte))

	projConf := project.Config{}
	if err := projConf.Load(name, db); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	branchConf := projConf.GetBranchByName(branchName)
	if branchConf == nil {
		branchConf = branch.NewConfig(branchName, projConf.Name, projConf.RepoName, globalConf.WorkSpace)
	}
	j, err := branchConf.ToJson()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.Write(j)
}
