package main

import (
	"bufio"
	"fmt"
	"github.com/distribyted/distribyted/http"
	"github.com/distribyted/distribyted/module"
	"github.com/distribyted/distribyted/webdav"
	"os"
	"os/signal"
	"path/filepath"
	"runtime"
	"syscall"
	"time"

	"github.com/anacrolix/missinggo/v2/filecache"
	"github.com/anacrolix/torrent/storage"
	"github.com/distribyted/distribyted/config"
	"github.com/rs/zerolog/log"
	"github.com/urfave/cli/v2"

	"github.com/distribyted/distribyted/fuse"
	dlog "github.com/distribyted/distribyted/log"
	_ "github.com/distribyted/distribyted/module"
	"github.com/distribyted/distribyted/torrent"
)

const (
	configFlag     = "config"
	fuseAllowOther = "fuse-allow-other"
	portFlag       = "http-port"
	webDAVPortFlag = "webdav-port"
)

func main() {
	app := &cli.App{
		Name:  "distribyted",
		Usage: "Torrent client with on-demand file downloading as a filesystem.",
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:    configFlag,
				Value:   "./distribyted-data/config/config.yaml",
				EnvVars: []string{"DISTRIBYTED_CONFIG"},
				Usage:   "YAML file containing distribyted configuration.",
			},
			&cli.IntFlag{
				Name:    portFlag,
				Value:   4444,
				EnvVars: []string{"DISTRIBYTED_HTTP_PORT"},
				Usage:   "HTTP port for web interface.",
			},
			&cli.IntFlag{
				Name:    webDAVPortFlag,
				Value:   36911,
				EnvVars: []string{"DISTRIBYTED_WEBDAV_PORT"},
				Usage:   "Port used for WebDAV interface.",
			},
			&cli.BoolFlag{
				Name:    fuseAllowOther,
				Value:   false,
				EnvVars: []string{"DISTRIBYTED_FUSE_ALLOW_OTHER"},
				Usage:   "Allow other users to access all fuse mountpoints. You need to add user_allow_other flag to /etc/fuse.conf file.",
			},
		},

		Action: func(c *cli.Context) error {
			err := load(c.String(configFlag), c.Int(portFlag), c.Int(webDAVPortFlag), c.Bool(fuseAllowOther))

			// stop program execution on errors to avoid flashing consoles
			if err != nil && runtime.GOOS == "windows" {
				log.Error().Err(err).Msg("problem starting application")
				fmt.Print("Press 'Enter' to continue...")
				bufio.NewReader(os.Stdin).ReadBytes('\n')
			}

			return err
		},

		HideHelpCommand: true,
	}

	if err := app.Run(os.Args); err != nil {
		log.Fatal().Err(err).Msg("problem starting application")
	}
}

func load(configPath string, port, webDAVPort int, fuseAllowOther bool) error {
	// configPath=./distribyted-data/config/config.yaml
	ch := config.NewHandler(configPath)

	conf, err := ch.Get()
	if err != nil {
		return fmt.Errorf("error loading configuration: %w", err)
	}
	// 日志加载
	dlog.Load(conf.Log)
	// 创建metadata文件夹
	if err := os.MkdirAll(conf.Torrent.MetadataFolder, 0744); err != nil {
		return fmt.Errorf("error creating metadata folder: %w", err)
	}

	cf := filepath.Join(conf.Torrent.MetadataFolder, "cache")
	fc, err := filecache.NewCache(cf) // 文件缓存
	if err != nil {
		return fmt.Errorf("error creating cache: %w", err)
	}

	st := storage.NewResourcePieces(fc.AsResourceProvider())

	// cache is not working with windows
	if runtime.GOOS == "windows" {
		st = storage.NewFile(cf)
	}
	// metadata-item建badger实例
	fis, err := torrent.NewFileItemStore(filepath.Join(conf.Torrent.MetadataFolder, "items"), 2*time.Hour)
	if err != nil {
		return fmt.Errorf("error starting item store: %w", err)
	}
	// 实例化torrent客户端
	c, err := torrent.NewClient(st, fis, conf.Torrent)
	if err != nil {
		return fmt.Errorf("error starting torrent client: %w", err)
	}

	pcp := filepath.Join(conf.Torrent.MetadataFolder, "piece-completion")
	if err := os.MkdirAll(pcp, 0744); err != nil {
		return fmt.Errorf("error creating piece completion folder: %w", err)
	}

	pc, err := storage.NewBoltPieceCompletion(pcp)
	if err != nil {
		return fmt.Errorf("error creating servers piece completion: %w", err)
	}

	var servers []*torrent.Server // 实例化
	for _, s := range conf.Servers {
		server := torrent.NewServer(c, pc, s)
		servers = append(servers, server)
		if err := server.Start(); err != nil {
			return fmt.Errorf("error starting server: %w", err)
		}
	}

	// 初始化badger数据库
	if err = module.InitBadger(conf); err != nil {
		return err
	}

	mh := fuse.NewHandler(fuseAllowOther || conf.Fuse.AllowOther, conf.Fuse.Path)

	sigChan := make(chan os.Signal)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		// 退出信号，关闭缓存、数据库等
		<-sigChan
		log.Info().Msg("closing servers...")
		for _, s := range servers {
			if err := s.Close(); err != nil {
				log.Warn().Err(err).Msg("problem closing server")
			}
		}
		log.Info().Msg("closing items database...")
		fis.Close()
		log.Info().Msg("closing magnet database...")
		//dbl.Close()
		module.Badger.Close()
		log.Info().Msg("closing torrent client...")
		c.Close()
		//log.Info().Msg("closing redis client...")
		//module.RedisClient.Close()
		log.Info().Msg("unmounting fuse filesystem...")
		mh.Unmount()

		log.Info().Msg("exiting321")
		os.Exit(1)
	}()

	log.Info().Msg(fmt.Sprintf("setting cache size to %d MB", conf.Torrent.GlobalCacheSize))
	fc.SetCapacity(conf.Torrent.GlobalCacheSize * 1024 * 1024)

	go func() {
		if conf.WebDAV != nil {
			if webDAVPort != 0 {
				conf.WebDAV.Port = webDAVPort
				port = webDAVPort
			}
			if err := webdav.NewWebDAVServer1(conf, c); err != nil {
				log.Error().Err(err).Msg("error starting webDAV")
			}
		}

		log.Warn().Msg("webDAV configuration not found!")
	}()

	err = http.NewGin(conf)
	log.Error().Err(err).Msg("error initializing HTTP server")
	return err
}
