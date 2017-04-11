package up

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"runtime/debug"
	"strconv"

	myctx "github.com/dev-cloverlab/harbor/middleware/context"
	"github.com/dev-cloverlab/harbor/models/branch"
	"github.com/dev-cloverlab/harbor/models/harbor"
	myproj "github.com/dev-cloverlab/harbor/models/project"
	"github.com/docker/libcompose/docker"
	"github.com/docker/libcompose/docker/ctx"
	"github.com/docker/libcompose/project"
	"github.com/docker/libcompose/project/options"
	"github.com/syndtr/goleveldb/leveldb"
)

const (
	EnvKeyPort1 = "PORT1"
	EnvKeyPort2 = "PORT2"
	EnvKeyPort3 = "PORT3"
	EnvKeyPort4 = "PORT4"
	EnvKeyPort5 = "PORT5"
)

type Request struct {
	Name         string `json:"name"`
	SourceBranch string `json:"branch"`
}

func (m Request) IsValid() error {
	if m.Name == "" {
		return fmt.Errorf("name is required")
	}
	if m.SourceBranch == "" {
		return fmt.Errorf("branch is required")
	}
	return nil
}

func Handler(db *leveldb.DB, w http.ResponseWriter, r *http.Request) {

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

	// load global settings from request context
	globalConf := &myctx.Config{}
	globalConf.FromJson(r.Context().Value(myctx.ConfigKey).([]byte))

	// load harbor settings from leveldb
	harborConf := &harbor.Config{}
	if err := harborConf.Load(db); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// load project configuration data from leveldb
	projConf := &myproj.Config{}
	if err := projConf.Load(req.Name, db); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	var branchConf *branch.Config

	if !projConf.KnownBranch(req.SourceBranch) {

		ports, err := harborConf.AllocPorts()
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		harborConf.Save(db)
		fmt.Println(ports)
		projConf.AddBranch(req.SourceBranch, projConf.Name, projConf.RepoName, globalConf.WorkSpace)
		branchConf = projConf.GetBranchByName(req.SourceBranch)
		branchConf.AllocPorts = ports

	} else {

		branchConf = projConf.GetBranchByName(req.SourceBranch)
		if branchConf.State == branch.DeployStateStart {
			bj, err := branchConf.ToJson()
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
			w.Write(bj)
			return
		}
	}

	branchConf.State = branch.DeployStateStart
	if err := projConf.Save(db); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	go func() {

		defer func() {
			err := recover()
			if err != nil {
				branchConf.Notice = err
				log.Println(err)
				debug.PrintStack()
			} else {
				branchConf.Notice = "Deploy Successfully"
			}
			branchConf.State = branch.DeployStateDone
			branchConf.DeployedAt = globalConf.Timestamp
			if err := projConf.Save(db); err != nil {
				panic(err)
			}
		}()

		// chdir
		if err := os.Chdir(branchConf.WorkDir); err != nil {
			panic(err)
		}

		// docker-compose up --build
		for i, port := range []string{EnvKeyPort1, EnvKeyPort2, EnvKeyPort3, EnvKeyPort4, EnvKeyPort5} {
			os.Setenv(port, strconv.Itoa(branchConf.AllocPorts[i]))
		}
		project, err := docker.NewProject(&ctx.Context{
			Context: project.Context{
				ComposeFiles: []string{branchConf.GetComposeFilePath()},
			},
		}, nil)
		if err != nil {
			panic(err)
		}

		if err := project.Up(context.Background(), options.Up{}); err != nil {
			panic(err)
		}
	}()

	bj, err := branchConf.ToJson()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Write(bj)
}
