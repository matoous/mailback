package main

import (
	"context"

	"go.uber.org/zap"

	"github.com/matoous/mailback/internal/cfg"
	"github.com/matoous/mailback/internal/sender"
	"github.com/matoous/mailback/internal/store"
)

func main() {
	var sCfg cfg.SenderConfig
	var storageCfg cfg.StorageConfig
	if err := cfg.LoadConfigs(&sCfg, &storageCfg); err != nil {
		panic(err)
	}

	db, err := store.NewSQLiteStore("test.db")
	if err != nil {
		panic("failed to connect database")
	}
	defer db.Close()

	log, err := zap.NewDevelopment()
	if err != nil {
		panic("failed to init logger")
	}

	be := sender.New(db, log, &sCfg)
	be.Run(context.TODO())
}
