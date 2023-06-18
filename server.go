package oauth2

import (
	"context"
	"github.com/bytedance/sonic"
	"github.com/go-oauth2/oauth2/v4/errors"
	"github.com/go-oauth2/oauth2/v4/generates"
	"github.com/go-oauth2/oauth2/v4/manage"
	"github.com/go-oauth2/oauth2/v4/models"
	"github.com/go-oauth2/oauth2/v4/server"
	"github.com/go-oauth2/oauth2/v4/store"
	oredis "github.com/go-oauth2/redis/v4"
	"github.com/go-redis/redis/v8"
	sredis "github.com/go-session/redis/v3"
	"github.com/go-session/session/v3"
	"net/http"
)

type Server struct {
	*manage.Manager
	*server.Server
}

// NewServer 创建oauth2 server
func NewServer(conf RedisConf) *Server {
	initSession(conf)
	// 1. create to default authorization management instance.
	manager := manage.NewDefaultManager()
	// 1.1 set the authorization code grant token config.
	manager.SetAuthorizeCodeTokenCfg(DefaultAuthorizeCodeTokenCfg)
	// 1.2 set the refresh token config.
	manager.SetRefreshTokenCfg(DefaultRefreshTokenCfg)
	// 1.3 mapping the token store interface, set token store (redis).
	if conf.Type == ClusterType {
		manager.MapTokenStorage(oredis.NewRedisClusterStore(&redis.ClusterOptions{Addrs: conf.Addrs, Password: conf.Pass}, DefaultTokenStoragePrefixKey))
	} else {
		manager.MapTokenStorage(oredis.NewRedisStore(&redis.Options{Addr: conf.Addrs[0], Password: conf.Pass}, DefaultTokenStoragePrefixKey))
	}
	// 1.4 mapping the access token generate interface, generate access_token.
	manager.MapAccessGenerate(generates.NewAccessGenerate())
	// 1.5 mapping the client store interface, set client store.
	clientStore := store.NewClientStore()
	clientId, clientSecret, domain := "123456", "abcdef", "localhost:8080"
	clientStore.Set("123456", &models.Client{
		ID:     clientId,
		Secret: clientSecret,
		Domain: domain,
	})
	manager.MapClientStorage(clientStore)

	// 2. create authorization server.
	config := server.NewConfig()
	config.AllowGetAccessRequest = true
	server := server.NewServer(config, manager)

	// server needs to implement UserAuthorizationHandler and PasswordAuthorizationHandler.
	// 2.1 set the password authorization handler(get user id from username and password).
	server.SetPasswordAuthorizationHandler(func(ctx context.Context, clientID, username, password string) (userId string, err error) {
		return "user_id", nil
	})

	// 2.2 set the user authorization handler(get user id from request).
	server.SetUserAuthorizationHandler(func(w http.ResponseWriter, r *http.Request) (userId string, err error) {
		store, err := session.Start(r.Context(), w, r)
		if err != nil {
			return "", err
		}

		// 2.3.1 it's from /oauth/authorize?client_id=xxx and need to redirect to /login.
		if r.Method == "GET" {
			marshal, err := sonic.Marshal(r.Form)
			if err != nil {
				return "", err
			}
			store.Set(DefaultAuthorizeForm, string(marshal))
			store.Save()
			w.Header().Set("Location", DefaultLoginPageUrl)
			w.WriteHeader(http.StatusFound)
			return "", nil
		}

		// 2.3.2 it's allow auth and from /login redirect /oauth/authorize, will get code to redirect_uri.
		if r.Method == "POST" {
			if r.Form == nil {
				err = r.ParseForm()
				if err != nil {
					return "", err
				}
			}

			userID, ok := store.Get(DefaultLoginUserId)
			if !ok {
				// not userID in session, redirect to /login.
				w.Header().Set("Location", DefaultLoginPageUrl)
				w.WriteHeader(http.StatusFound)
				return
			}
			store.Delete(DefaultLoginUserId)
			store.Save()

			return userID.(string), nil
		}
		return "", nil
	})
	// 2.3 set client info handler
	server.SetClientInfoHandler(func(r *http.Request) (clientID, clientSecret string, err error) {
		// 从请求头获取clientId、clientSecret等
		clientID = r.FormValue("client_id")
		if len(clientID) == 0 {
			return "", "", errors.ErrInvalidRequest
		}

		clientSecret = r.FormValue("client_secret")
		if len(clientSecret) == 0 {
			return "", "", errors.ErrInvalidRequest
		}
		return
	})

	return &Server{Manager: manager, Server: server}
}

// use redis as session manager.
func initSession(conf RedisConf) {
	var store session.ManagerStore
	if conf.Type == ClusterType {
		store = sredis.NewRedisClusterStore(&sredis.ClusterOptions{Addrs: conf.Addrs, Password: conf.Pass}, DefaultRedisStorePrefixKey)
	} else {
		store = sredis.NewRedisStore(&sredis.Options{Addr: conf.Addrs[0], Password: conf.Pass}, DefaultRedisStorePrefixKey)
	}

	session.InitManager(session.SetStore(store))
}
