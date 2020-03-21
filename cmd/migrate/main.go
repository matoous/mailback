package main

import (
	"os"

	"go.uber.org/zap"

	"github.com/matoous/mailback/internal/cfg"
	"github.com/matoous/mailback/internal/store"
)

func main() {
	var storageCfg cfg.StorageConfig
	if err := cfg.LoadConfigs(&storageCfg); err != nil {
		panic(err)
	}

	log, err := zap.NewDevelopment()
	if err != nil {
		panic("failed to init logger")
	}
	defer func() {
		err := log.Sync()
		if err != nil {
			panic(err)
		}
	}()

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
	if err := db.Migrate(); err != nil {
		log.Error("storage.migrate", zap.Error(err))
		os.Exit(1)
	}
}
