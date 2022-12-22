package main

import (
	"os"

	"github.com/dezhishen/go-nitter-proxy/pkg/provider"
	"github.com/dezhishen/go-nitter-proxy/pkg/proxy"
	"github.com/gin-gonic/gin"
)

func main() {
	hostprovider := provider.NewGithubProvider()
	cfg := map[string]interface{}{
		"repoProxy": os.Getenv("REPO_PROXY"),
	}
	hostprovider.Init(cfg)
	proxy.Init(hostprovider)
	gin.SetMode(gin.ReleaseMode)
	r := gin.Default()
	r.Any(
		"/*proxyPath",
		func(c *gin.Context) {
			// 重定向
			// 获取重定向地址
			redirectUrl := proxy.ForwardHandler(c.Param("proxyPath"))
			c.Redirect(302, redirectUrl)
		},
	)
	r.Run(":8080")
}
