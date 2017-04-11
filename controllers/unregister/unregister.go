package unregister

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"

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
	Name string `json:"name"`
}

func (m Request) IsValid() error {
	if m.Name == "" {
		return fmt.Errorf("name is required")
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

	// load project configuration data
	projConf := myproj.Config{}
	if err := projConf.Load(req.Name, db); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// down all containers
	for _, branchConf := range projConf.Branch {
		project, err := docker.NewProject(&ctx.Context{
			Context: project.Context{
				ComposeFiles: []string{branchConf.GetComposeFilePath()},
			},
		}, nil)
		if err != nil {
			log.Println(err.Error())
			continue
		}

		if err := project.Down(context.Background(), options.Down{
			RemoveVolume: true,
		}); err != nil {
			log.Println(err.Error())
			continue
		}

		for _, port := range branchConf.AllocPorts {
			harborConf.ReleasePort(port)
		}
	}

	// remove project workspace
	dest := []string{
		conf.WorkSpace,
		projConf.RepoName,
	}
	if err := os.RemoveAll(strings.Join(dest, string(os.PathSeparator))); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// unregister project
	if err := db.Delete([]byte(req.Name), nil); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// restore harbor configuration data
	if err := harborConf.Save(db); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	j, err := harborConf.ToJson()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Write(j)
}
