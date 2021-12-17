package module

import (
	"fmt"
	"github.com/distribyted/distribyted/config"
	"github.com/distribyted/distribyted/torrent/loader"
	"path/filepath"
)

var Badger = &loader.DB{}

func InitBadger(conf *config.Root) error {
	// badger多个进程无法同时打开同一个数据库，初始化一个实例
	db, err := loader.NewDB(filepath.Join(conf.Torrent.MetadataFolder, "magnetdb"))
	if err != nil {
		// todo 日志
		return fmt.Errorf("error starting magnet database: %w", err)
	}
	Badger = db
	return nil
}
