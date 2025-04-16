package main

import (
	"log"
	"time"

	"g.pervovsky.ru/go-access-admin/internal/access"
	"g.pervovsky.ru/go-access-admin/internal/config"
	"g.pervovsky.ru/go-access-admin/internal/handler"
	"g.pervovsky.ru/go-access-admin/internal/scheduler"
	"g.pervovsky.ru/go-access-admin/internal/storage"
	"github.com/gin-gonic/gin"
)

func main() {
	cfg, err := config.LoadConfig("internal/config/config.yaml")
	if err != nil {
		log.Fatalf("error loading config: %v", err)
	}

	db, err := storage.NewDB("./accesses.db")
	if err != nil {
		log.Fatalf("error initializing db: %v", err)
	}
	defer db.Close()

	if cfg.AppSettings.SyncHtpasswd {
		if err := access.Synchronize(*db, *cfg); err != nil {
			log.Fatalf("error synchronizing db: %v", err)
		}
	}

	scheduler.StartCleaner(db, time.Duration(cfg.AppSettings.CleanAccessesInterval)*time.Minute)

	h := handler.Handler{
		DB:     db,
		Config: cfg,
	}

	r := gin.Default()
	h.RegisterRoutes(r)

	r.LoadHTMLFiles("./web/templates/index.html")
	r.Static("/static", "./web/static")

	log.Println("Server running on :8080")
	r.Run(":8080")
}
