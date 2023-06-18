package oauth2

import (
	"github.com/go-oauth2/oauth2/v4/manage"
	"time"
)

var (
	// ClusterType means redis cluster.
	ClusterType = "cluster"
	// NodeType means redis node.
	NodeType = "node"

	// DefaultAuthorizeCodeTokenCfg is the default authorization code grant token config.
	DefaultAuthorizeCodeTokenCfg = &manage.Config{AccessTokenExp: time.Hour * 24 * 7, RefreshTokenExp: time.Hour * 24 * 30, IsGenerateRefresh: true}
	// DefaultRefreshTokenCfg is the default refresh token config.
	DefaultRefreshTokenCfg = &manage.RefreshingConfig{IsGenerateRefresh: false, IsRemoveAccess: false, IsRemoveRefreshing: false}

	// DefaultTokenStoragePrefixKey is the default token storage prefix key.
	DefaultTokenStoragePrefixKey = "oauth2:token:"

	// DefaultRedisStorePrefixKey is the default session redis prefix key.
	DefaultRedisStorePrefixKey = "oauth2:store:"

	// DefaultAuthorizeForm is the default authorization form.
	DefaultAuthorizeForm = "DefaultAuthorizeForm"
	// DefaultLoginUserId is the default login user id.
	DefaultLoginUserId = "DefaultLoginUserId"

	DefaultLoginPageUrl = "/page/login"
	DefaultAuthPageUrl  = "/page/auth"
)

// RedisConf defines the redis configuration.
type RedisConf struct {
	Addrs []string `json:"addrs"`
	Pass  string   `json:"pass,optional"`
	Type  string   `json:"type,default=node"`
}
