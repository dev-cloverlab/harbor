package dbdriver

import (
	"net/http"

	"github.com/syndtr/goleveldb/leveldb"
)

type HandlerFunc func(db *leveldb.DB, w http.ResponseWriter, r *http.Request)

func Middleware(db *leveldb.DB, fn HandlerFunc) http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fn(db, w, r)
	})
}
