package main

import (
	"os"

	_ "github.com/jinzhu/gorm/dialects/sqlite"
	"go.uber.org/zap"

	"github.com/matoous/mailback/internal/cfg"
	"github.com/matoous/mailback/internal/receiver"
	"github.com/matoous/mailback/internal/store"
)

func main() {
	exitCode := 0
	defer func() {
		os.Exit(exitCode)
	}()
	var recCfg cfg.ReceiverConfig
	var storageCfg cfg.StorageConfig
	if err := cfg.LoadConfigs(&recCfg, &storageCfg); err != nil {
		panic(err)
	}

	log, err := zap.NewDevelopment()
	if err != nil {
		panic("failed to init logger")
	}

	db, err := store.NewSQLiteStore(storageCfg.Database)
	if err != nil {
		log.Error("store.init", zap.Error(err))
		exitCode++
		return
	}
	defer func() {
		err := db.Close()
		if err != nil {
			log.Error("store.close", zap.Error(err))
		}
	}()

	srv, err := receiver.New(db, log, recCfg)
	if err != nil {
		log.Error("receiver.init", zap.Error(err))
		exitCode++
		return
	}

	if err := srv.Run(); err != nil {
		log.Fatal("receiver.run", zap.Error(err))
	}
}
