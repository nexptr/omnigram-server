package selfhost

import (
	"context"
	"encoding/json"
	"os"
	"os/signal"
	"syscall"
	"testing"
	"time"

	"github.com/nexptr/llmchain/llms"
	"github.com/nexptr/omnigram-server/conf"
	"github.com/nexptr/omnigram-server/epub/schema"
	"github.com/nexptr/omnigram-server/log"
	"go.uber.org/zap/zapcore"
)

var basePath string

func init() {
	log.Init(`stdout`, zapcore.DebugLevel)
	testDir, _ := os.Getwd()

	basePath = testDir + `/../../`
}

func initStore() *schema.Store {

	opt := &schema.Opt{
		Driver:   schema.DRSQLite,
		LogLevel: zapcore.DebugLevel,
		Host:     basePath + "build/cxbooks.db",
	}
	log.I(`打开数据库连接...`)
	store, _ := schema.OpenDB(opt)
	log.I(`打开数据库连接`)
	return store
}

func TestScanBooks(t *testing.T) {

	cf := &conf.Config{
		APIAddr:      "",
		LogLevel:     0,
		LogDir:       "",
		ModelOptions: []llms.ModelOptions{},
		EpubOptions: conf.EpubOptions{
			DataPath:           basePath + `build/epub`,
			CachePath:          basePath + `build`,
			SaveCoverBesideSrc: false,
			MaxEpubSize:        0,
			DBConfig:           &schema.Opt{},
		},
	}

	ch := make(chan os.Signal, 1)
	signal.Notify(ch, os.Interrupt, syscall.SIGTERM)

	log.I(`初始化扫描管理`)

	kv, _ := schema.OpenLocalDir(basePath + `build`)

	manager, _ := NewScannerManager(context.TODO(), cf, kv, initStore())

	manager.Start(2)
	ticker := time.NewTicker(3 * time.Second)

	for {

		select {
		case <-ch:
			println(`exit.....`)
			return
		case <-ticker.C:
			states := manager.Status()

			str, _ := json.Marshal(states)

			println(string(str))

		}

	}

}
