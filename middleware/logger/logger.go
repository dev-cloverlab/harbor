package logger

import (
	"log"
	"net/http"
	"net/http/httptest"
	"runtime/debug"
	"time"
)

func Middleware(fn http.HandlerFunc) http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		begin := time.Now()

		rec := httptest.NewRecorder()
		fn.ServeHTTP(rec, r)

		for k, vl := range rec.Header() {
			for _, v := range vl {
				w.Header().Add(k, v)
			}
		}
		if rec.Code == 0 {
			rec.Code = http.StatusOK
		}

		errStr := ""
		if rec.Code != http.StatusOK {
			errStr = string(rec.Body.Bytes())
			w.WriteHeader(rec.Code)
			debug.PrintStack()
		}
		w.Write(rec.Body.Bytes())

		log.Printf(
			"elapsed:%f\tcode:%d\tmethod:%s\turi:%s\terr:%s",
			time.Since(begin).Seconds(),
			rec.Code,
			r.Method,
			r.URL.RequestURI(),
			errStr,
		)
	})
}
