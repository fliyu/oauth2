package main

import (
	"fmt"
	"github.com/bytedance/sonic"
	"github.com/go-session/session/v3"
	"net/http"
	"net/url"
	"oauth2"
	"oauth2/server/types"
	"oauth2/server/utils"
	"os"
)

var server *oauth2.Server

func init() {
	server = oauth2.NewServer(oauth2.RedisConf{Addrs: []string{"127.0.0.1:6379"}})
}

func main() {
	mux := http.NewServeMux()
	mux.HandleFunc("/oauth/authorize", authorizeHandler)
	mux.HandleFunc("/login", loginHandler)
	mux.HandleFunc("/oauth/access_token", accessTokenHandler)
	mux.HandleFunc("/oauth/refresh_token", refreshTokenHandler)

	mux.HandleFunc(oauth2.DefaultLoginPageUrl, loginPageHandler)
	mux.HandleFunc(oauth2.DefaultAuthPageUrl, authPageHandler)
	server := http.Server{
		Addr:    ":8088",
		Handler: mux,
	}
	fmt.Println("Start server at http://localhost:8088/")
	server.ListenAndServe()
}

// 授权接口
func authorizeHandler(w http.ResponseWriter, r *http.Request) {
	// There will be two steps here, one is to enter through the /authorize interface,
	// and the other is to enter when obtaining the code
	store, err := session.Start(r.Context(), w, r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	var form url.Values
	v, ok := store.Get(oauth2.DefaultAuthorizeForm)
	if !ok {
		// oauth2 access GET /oauth/authorize?client_id=123456&redirect_uri=http://localhost:8080/redirect.html&response_type=code&scope=scope
		// Do some verification, such as client_id, redirect_uri, etc.
		var req types.AuthorizeReq
		err = utils.Parse(r, &req)
		if err != nil {
			return
		}
	} else {
		// oauth2 access POST /oauth/authorize get code to redirect_uri
		err = sonic.Unmarshal([]byte(v.(string)), &form)
		if err != nil {
			return
		}
	}
	r.Form = form

	store.Delete(oauth2.DefaultAuthorizeForm)
	store.Save()

	err = server.HandleAuthorizeRequest(w, r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
	}
}

// 登录接口，校验用户名密码
func loginHandler(w http.ResponseWriter, r *http.Request) {
	var req types.LoginReq
	err := utils.Parse(r, &req)
	if err != nil {
		return
	}

	store, err := session.Start(r.Context(), w, r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if r.Form == nil {
		if err := r.ParseForm(); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	}

	if req.Username != "admin" || req.Password != "123456" {
		w.WriteHeader(http.StatusForbidden)
		return
	}

	store.Set(oauth2.DefaultLoginUserId, r.Form.Get("username"))
	store.Save()

	w.Header().Set("Location", oauth2.DefaultAuthPageUrl)
	w.WriteHeader(http.StatusFound)
}

// 获取access_token
func accessTokenHandler(w http.ResponseWriter, r *http.Request) {
	var req types.AccessTokenReq
	err := utils.Parse(r, &req)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	if err := r.ParseForm(); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	form := url.Values{}
	form.Set("client_id", req.ClientId)
	form.Set("client_secret", req.ClientSecret)
	form.Set("code", req.Code)
	form.Set("grant_type", req.GrantType)
	form.Set("redirect_uri", req.RedirectUri)
	r.Form = form
	err = server.HandleTokenRequest(w, r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
}

// 刷新access_token
func refreshTokenHandler(w http.ResponseWriter, r *http.Request) {
	var req types.RefreshTokenReq
	err := utils.Parse(r, &req)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	if err := r.ParseForm(); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	form := url.Values{}
	form.Set("client_id", req.ClientId)
	form.Set("client_secret", req.ClientSecret)
	form.Set("refresh_token", req.RefreshToken)
	form.Set("grant_type", req.GrantType)
	form.Set("redirect_uri", req.RedirectUri)
	r.Form = form
	err = server.HandleTokenRequest(w, r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
}

// 登录页面
func loginPageHandler(w http.ResponseWriter, r *http.Request) {
	outputHTML(w, r, "static/login.html")
}

// 授权页面
func authPageHandler(w http.ResponseWriter, r *http.Request) {
	outputHTML(w, r, "static/auth.html")
}

func outputHTML(w http.ResponseWriter, req *http.Request, filename string) {
	file, err := os.Open(filename)
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}
	defer file.Close()
	fi, _ := file.Stat()
	http.ServeContent(w, req, file.Name(), fi.ModTime(), file)
}
