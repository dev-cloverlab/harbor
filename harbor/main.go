package main

import (
	"context"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"regexp"

	"github.com/dev-cloverlab/harbor/controllers/br"
	"github.com/dev-cloverlab/harbor/controllers/down"
	"github.com/dev-cloverlab/harbor/controllers/list"
	"github.com/dev-cloverlab/harbor/controllers/register"
	"github.com/dev-cloverlab/harbor/controllers/unregister"
	"github.com/dev-cloverlab/harbor/controllers/up"
	myctx "github.com/dev-cloverlab/harbor/middleware/context"
	"github.com/dev-cloverlab/harbor/middleware/dbdriver"
	"github.com/dev-cloverlab/harbor/middleware/header"
	"github.com/dev-cloverlab/harbor/middleware/logger"
	"github.com/dev-cloverlab/harbor/models/harbor"
	"github.com/syndtr/goleveldb/leveldb"
)

var proxyRegexp *regexp.Regexp = regexp.MustCompile("proxy")

//go:generate go-assets-builder -s="/views/build" -o bindata.go ../views/build

func route(db *leveldb.DB, conf myctx.Config) {
	http.HandleFunc("/", logger.Middleware(func(w http.ResponseWriter, r *http.Request) {
		f, err := Assets.Open("/index.html")
		if err != nil {
			http.Error(w, err.Error(), http.StatusNotFound)
			return
		}
		buf, err := ioutil.ReadAll(f)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		w.Write(buf)
	}))
	http.Handle("/static/", http.FileServer(Assets))
	http.HandleFunc("/register", logger.Middleware(header.Middleware(myctx.Middleware(conf, dbdriver.Middleware(db, register.Handler)))))
	http.HandleFunc("/unregister", logger.Middleware(header.Middleware(myctx.Middleware(conf, dbdriver.Middleware(db, unregister.Handler)))))
	http.HandleFunc("/up", logger.Middleware(header.Middleware(myctx.Middleware(conf, dbdriver.Middleware(db, up.Handler)))))
	http.HandleFunc("/down", logger.Middleware(header.Middleware(myctx.Middleware(conf, dbdriver.Middleware(db, down.Handler)))))
	http.HandleFunc("/list", logger.Middleware(header.Middleware(myctx.Middleware(conf, dbdriver.Middleware(db, list.Handler)))))
	http.HandleFunc("/br", logger.Middleware(header.Middleware(myctx.Middleware(conf, dbdriver.Middleware(db, br.Handler)))))
}

func main() {

	port := flag.String("p", ":8080", `server listen port`)
	ws := flag.String("w", "./harbor", `application workspace path`)
	portRange := flag.String("r", "10000:12000", `port range for containers`)

	flag.Parse()

	db, err := leveldb.OpenFile(fmt.Sprintf("%s%sdatabase", *ws, string(os.PathSeparator)), nil)
	if err != nil {
		panic(err)
	}
	defer db.Close()

	harborConf := harbor.Config{}
	if err := harborConf.Load(db); err != nil {
		harborConf.PortsAllocated = []int{}
	}
	harborConf.PortRange = harbor.PortRange(*portRange)
	harborConf.Save(db)

	conf := myctx.Config{
		WorkSpace: *ws,
		Timestamp: time.Now(),
	}
	route(db, conf)

	srv := &http.Server{Addr: *port}
	go func() {
		log.Printf("Start server %s\n", *port)
		if err := srv.ListenAndServe(); err != nil {
			log.Println(err.Error())
		}
	}()

	sigCh := make(chan os.Signal, 1)
	signal.Notify(
		sigCh,
		syscall.SIGTERM,
		syscall.SIGINT,
	)
	<-sigCh

	ctx, _ := context.WithTimeout(context.Background(), 5*time.Second)
	if err := srv.Shutdown(ctx); err != nil {
		panic(err)
	}
	log.Println("Shutdown collectry")
}
