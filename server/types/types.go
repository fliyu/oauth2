package types

// AuthorizeReq 用户授权请求，如果用户同意授权，页面将跳转至 redirect_uri/?code=CODE&state=STATE。
type AuthorizeReq struct {
	ClientId     string `query:"client_id"`
	RedirectUri  string `query:"redirect_uri"`
	ResponseType string `query:"response_type,default=code"`
	Scope        string `query:"scope,default=scope"`
	State        string `query:"state,optional"`
}
type AuthorizeResp struct {
}

// LoginReq 登录请求
type LoginReq struct {
	Username string `form:"username"`
	Password string `form:"password"`
}
type LoginResp struct {
}

// AccessTokenReq 通过code换取access_token
type AccessTokenReq struct {
	ClientId     string `json:"client_id"`
	ClientSecret string `json:"client_secret"`
	Code         string `json:"code"`
	GrantType    string `json:"grant_type,default=authorization_code"`
	RedirectUri  string `json:"redirect_uri"`
}
type AccessTokenRsp struct {
}

// RefreshTokenReq 刷新access_token
type RefreshTokenReq struct {
	ClientId     string `json:"client_id"`
	ClientSecret string `json:"client_secret"`
	RefreshToken string `json:"refresh_token"`
	GrantType    string `json:"grant_type,default=refresh_token"`
	RedirectUri  string `json:"redirect_uri"`
}
type RefreshTokenRsp struct {
}
