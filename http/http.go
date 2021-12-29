package http

import (
	"fmt"
	"github.com/distribyted/distribyted/controller"
	"net/http"

	"github.com/anacrolix/missinggo/v2/filecache"
	"github.com/distribyted/distribyted"
	"github.com/distribyted/distribyted/config"
	"github.com/distribyted/distribyted/torrent"
	"github.com/gin-gonic/gin"
	"github.com/rs/zerolog/log"
	"github.com/shurcooL/httpfs/html/vfstemplate"
)

func New(fc *filecache.Cache, ss *torrent.Stats, s *torrent.Service, ch *config.Handler, tss []*torrent.Server, fs http.FileSystem, logPath string, cfg *config.HTTPGlobal) error {
	gin.SetMode(gin.ReleaseMode)
	r := gin.New()
	r.Use(gin.Recovery())
	r.Use(gin.ErrorLogger())

	r.GET("/assets/*filepath", func(c *gin.Context) {
		c.FileFromFS(c.Request.URL.Path, http.FS(distribyted.Assets))
	})

	if cfg.HTTPFS {
		log.Info().Str("host", fmt.Sprintf("%s:%d/fs", cfg.IP, cfg.Port)).Msg("starting HTTPFS")
		r.GET("/fs/*filepath", func(c *gin.Context) {
			path := c.Param("filepath")
			c.FileFromFS(path, fs)
		})
	}

	t, err := vfstemplate.ParseGlob(http.FS(distribyted.Templates), nil, "/templates/*")
	if err != nil {
		return fmt.Errorf("error parsing html: %w", err)
	}

	r.SetHTMLTemplate(t)

	r.GET("/", indexHandler)
	r.GET("/login", loginHandler)
	r.GET("/routes", routesHandler(ss))
	r.GET("/logs", logsHandler)
	r.GET("/servers", serversFoldersHandler())

	api := r.Group("/api")
	{
		api.GET("/log", apiLogHandler(logPath))
		api.GET("/status", apiStatusHandler(fc, ss))
		api.GET("/servers", apiServersHandler(tss))

		api.GET("/routes", apiRoutesHandler(ss))
		api.POST("/routes/:route/torrent", apiAddTorrentHandler(s))
		api.DELETE("/routes/:route/torrent/:torrent_hash", apiDelTorrentHandler(s))

	}

	log.Info().Str("host", fmt.Sprintf("%s:%d", cfg.IP, cfg.Port)).Msg("starting webserver")

	if err := r.Run(fmt.Sprintf("%s:%d", cfg.IP, cfg.Port)); err != nil {
		return fmt.Errorf("error initializing server: %w", err)
	}

	return nil
}

func NewGin(conf *config.Root) error {
	router := gin.Default()
	// noAuth：无需登录
	noAuth := router.Group("/v1")
	{
		noAuth.POST("/login", controller.UserLogin)
		noAuth.GET("/user/info", controller.UserInfo)
		//noAuth.POST("/read", readEndpoint)
	}

	// auth：需要登录
	//auth := router.Group("/v2")
	//{
	//	auth.POST("/login", loginEndpoint)
	//	auth.POST("/submit", submitEndpoint)
	//	auth.POST("/read", readEndpoint)
	//}

	router.Run(":4444")
	return nil
}
