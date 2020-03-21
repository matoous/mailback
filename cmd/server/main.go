package main

import (
	"os"

	_ "github.com/jinzhu/gorm/dialects/sqlite"
	"go.uber.org/zap"

	"github.com/matoous/mailback/internal/cfg"
	"github.com/matoous/mailback/internal/server"
	"github.com/matoous/mailback/internal/store"
)

func main() {
	var webConfig cfg.WebServerConfig
	var storageCfg cfg.StorageConfig
	if err := cfg.LoadConfigs(&webConfig, &storageCfg); err != nil {
		panic(err)
	}

	log, err := zap.NewDevelopment()
	if err != nil {
		panic("failed to init logger")
	}

	db, err := store.NewSQLiteStore(storageCfg.Database)
	if err != nil {
		log.Error("storage.init", zap.Error(err))
		os.Exit(1)
	}
	defer func() {
		err := db.Close()
		if err != nil {
			log.Error("storage.close", zap.Error(err))
		}
	}()

	s, err := server.New(db, log, webConfig)
	if err != nil {
		log.Error("server.init", zap.Error(err))
		os.Exit(1)
	}

	if err := s.Run(); err != nil {
		log.Error("server.run", zap.Error(err))
		os.Exit(1)
	}
}
