package down

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"

	myctx "github.com/dev-cloverlab/harbor/middleware/context"
	"github.com/dev-cloverlab/harbor/models/harbor"
	myproj "github.com/dev-cloverlab/harbor/models/project"

	"github.com/docker/libcompose/docker"
	"github.com/docker/libcompose/docker/ctx"
	"github.com/docker/libcompose/project"
	"github.com/docker/libcompose/project/options"
	"github.com/syndtr/goleveldb/leveldb"
)

type Request struct {
	Name   string `json:"name"`
	Branch string `json:"branch"`
}

func (m Request) IsValid() error {
	if m.Name == "" {
		return fmt.Errorf("name is required")
	}
	if m.Branch == "" {
		return fmt.Errorf("branch is required")
	}
	return nil
}

func Handler(db *leveldb.DB, w http.ResponseWriter, r *http.Request) {

	// load global settings from request context
	conf := &myctx.Config{}
	conf.FromJson(r.Context().Value(myctx.ConfigKey).([]byte))

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

	// load harbor settings from leveldb
	harborConf := &harbor.Config{}
	if err := harborConf.Load(db); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// load project data
	projConf := myproj.Config{}
	if err := projConf.Load(req.Name, db); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// down containers
	branchConf := projConf.GetBranchByName(req.Branch)
	if branchConf == nil {
		log.Println("branch is not found")
	}

	if branchConf != nil {
		project, err := docker.NewProject(&ctx.Context{
			Context: project.Context{
				ComposeFiles: []string{branchConf.GetComposeFilePath()},
			},
		}, nil)
		if err != nil {
			log.Println(err.Error())
		}
		if project != nil {
			if err := project.Down(context.Background(), options.Down{
				RemoveVolume: true,
			}); err != nil {
				log.Println(err.Error())
			}
		}

		// remove project branch
		if err := os.RemoveAll(branchConf.WorkDir); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		for _, port := range branchConf.AllocPorts {
			harborConf.ReleasePort(port)
		}
	}

	// delete branch from project
	projConf.DelBranch(req.Branch)

	// restore project configuration data
	if err := projConf.Save(db); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// restore harbor configuration data
	if err := harborConf.Save(db); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	j, err := projConf.ToJson()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.Write(j)
}
