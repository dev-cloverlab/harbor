package list

import (
	"encoding/json"
	"net/http"

	"github.com/dev-cloverlab/harbor/models/harbor"
	"github.com/dev-cloverlab/harbor/models/project"
	"github.com/syndtr/goleveldb/leveldb"
)

type Response struct {
	Projects []*project.Config `json:"projects"`
}

func Handler(db *leveldb.DB, w http.ResponseWriter, r *http.Request) {

	projConf := []*project.Config{}

	iter := db.NewIterator(nil, nil)
	defer iter.Release()
	for iter.Next() {
		c := &project.Config{}
		if string(iter.Key()) == harbor.ConfigKey {
			continue
		}
		if err := c.Load(string(iter.Key()), db); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		projConf = append(projConf, c)
	}
	if err := iter.Error(); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	res := Response{
		Projects: projConf,
	}
	j, err := json.Marshal(res)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.Write(j)
}
