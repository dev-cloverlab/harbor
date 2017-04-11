package context

import (
	"context"
	"encoding/json"
	"net/http"
	"time"
)

const (
	ConfigKey string = "config"
)

type Config struct {
	WorkSpace string    `json:"workspace"`
	Timestamp time.Time `json:"timestamp"`
}

func (m *Config) FromJson(j []byte) error {
	return json.Unmarshal(j, m)
}

func Middleware(conf Config, fn http.HandlerFunc) http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		buf, err := json.Marshal(conf)
		if err != nil {
			panic(err)
		}

		ctx := r.Context()
		ctx = context.WithValue(ctx, ConfigKey, buf)
		cr := r.WithContext(ctx)
		fn.ServeHTTP(w, cr)
	})
}
