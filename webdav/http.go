package webdav

import (
	"errors"
	"fmt"
	"github.com/anacrolix/torrent"
	"github.com/distribyted/distribyted/config"
	"github.com/distribyted/distribyted/fuse"
	"github.com/distribyted/distribyted/module"
	mTorrent "github.com/distribyted/distribyted/torrent"
	"github.com/distribyted/distribyted/torrent/loader"
	"golang.org/x/net/webdav"
	"net/http"
	"time"
	"unsafe"

	"github.com/distribyted/distribyted/fs"
	"github.com/rs/zerolog/log"
)

const fileSystemCacheKey = "file_system_user"

var globalWebdav = make(map[string]*webdav.Handler)

func NewWebDAVServer(fs fs.Filesystem, port int, user, pass string) error {
	log.Info().Str("host", fmt.Sprintf("0.0.0.0:%d", port)).Msg("starting webDAV server")

	srv := newHandler(fs)

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		username, password, _ := r.BasicAuth()
		if username == user && password == pass {
			srv.ServeHTTP(w, r)
			return
		}

		w.Header().Set("WWW-Authenticate", `Basic realm="BASIC WebDAV REALM"`)
		w.WriteHeader(401)
		w.Write([]byte("401 Unauthorized\n"))
	})

	return http.ListenAndServe(fmt.Sprintf("0.0.0.0:%d", port), nil)
}

func NewWebDAVServer1(conf *config.Root, c *torrent.Client) error {
	log.Info().Str("host", fmt.Sprintf("0.0.0.0:%d", conf.WebDAV.Port)).Msg("starting webDAV server")

	//dbl, err := loader.NewDB(filepath.Join(conf.Torrent.MetadataFolder, "magnetdb"))
	//if err != nil {
	//	// todo 日志
	//	return fmt.Errorf("error starting magnet database: %w", err)
	//}
	//defer dbl.Close()

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		username, password, _ := r.BasicAuth()
		fmt.Print(username)
		if username != "" {
			if !validUser(username, password) {
				goto unAuth
			}
			// todo 先简单处理一下账号权限目录
			srv, _ := doHttp(username, conf, c)
			srv.ServeHTTP(w, r)
			return
		}
	unAuth:
		w.Header().Set("WWW-Authenticate", `Basic realm="BASIC WebDAV REALM"`)
		w.WriteHeader(401)
		w.Write([]byte("401 Unauthorized\n"))
	})

	return http.ListenAndServe(fmt.Sprintf("0.0.0.0:%d", conf.WebDAV.Port), nil)
}

// 校验用户
func validUser(user, password string) bool {
	resPw, err := module.Badger.GetUserPassword(user)
	if err != nil {
		// todo 日志
	}
	if resPw != password {
		return false
	}
	return true
}

func doHttp(user string, conf *config.Root, c *torrent.Client) (*webdav.Handler, error) {
	var (
		err error
	)

	if srv, ok := globalWebdav[user]; ok {
		return srv, nil
	}

	var routes []*config.Route
	for _, routeV := range conf.Routes {
		if routeV.Name != user {
			continue
		}
		routes = append(routes, routeV)
	}
	cl := loader.NewConfig(routes)
	ss := mTorrent.NewStats()
	tus := mTorrent.NewService(cl, module.Badger, ss, c, conf.Torrent.AddTimeout, user)
	fss, err := tus.Load()
	if err != nil {
		return nil, fmt.Errorf("error when loading torrents: %w", err)
	}

	mh := fuse.NewHandler(conf.Fuse.AllowOther, conf.Fuse.Path)
	log.Info().Msg(fmt.Sprintf("当前mount用户：%s", user))
	if err := mh.Mount(fss); err != nil {
		log.Info().Err(err).Msg("error mounting filesystems")
	}

	cfs, err := fs.NewContainerFs(fss)
	if err != nil {
		log.Error().Err(err).Msg("error adding files to webDAV")
		return nil, errors.New("error adding files to webDAV")
	}
	srv := newHandler(cfs)
	// 写入缓存
	//setUserFsCache(fss, user)
	globalWebdav[user] = srv
	return srv, nil

}

// 设置用户file_system缓存
func setUserFsCache(fss map[string]fs.Filesystem, user string) {
	// todo 为空判断
	type SliceMock struct {
		addr uintptr
		len  int
		cap  int
	}
	asd := &fss
	Len := unsafe.Sizeof(*asd)
	srvBytes := &SliceMock{
		addr: uintptr(unsafe.Pointer(asd)),
		cap:  int(Len),
		len:  int(Len),
	}
	byteData := *(*[]byte)(unsafe.Pointer(srvBytes))
	fmt.Println("string:", string(byteData))
	key := fmt.Sprintf("%s_%s", fileSystemCacheKey, user)
	module.RedisClient.Set(key, string(byteData), 24*time.Hour)
}

// 获取用户file_system缓存
func getUserFsCache(user string) (map[string]fs.Filesystem, error) {
	key := fmt.Sprintf("%s_%s", fileSystemCacheKey, user)
	str, err := module.RedisClient.Get(key).Result()
	if err != nil {
		return nil, err
	}
	if str == "" {
		return nil, errors.New(fmt.Sprintf("%s is empty", key))
	}
	byteData := []byte(str)
	var data = *(**map[string]fs.Filesystem)(unsafe.Pointer(&byteData))
	module.RedisClient.Set(key, str, 24*time.Hour)
	return *data, nil
}

func Lock1() bool {

	for {

		resp := module.RedisClient.SetNX("asd123", 1, 0) //返回执行结果

		lockSuccess, err := resp.Result()

		if err == nil && lockSuccess {

			fmt.Println("lock success!")

			return true

		} else {

			//抢锁失败，继续自旋
			time.Sleep(10 * time.Millisecond)

			fmt.Println("lock failed!", err)

		}

	}

}

func Unlock1() {

	delResp := module.RedisClient.Del("asd123")

	unlockSuccess, err := delResp.Result()

	if err == nil && unlockSuccess > 0 {

		fmt.Println("unlock success!")

	} else {

		fmt.Println("unlock failed!", err)

	}

}
