package proxy

import (
	"log"
	"net/url"

	"github.com/dezhishen/go-nitter-proxy/pkg/provider"
)

var hostprovider provider.HostProvider

func Init(_provider provider.HostProvider) {
	hostprovider = _provider
}

func ForwardHandler(path string) string {
	host, err := hostprovider.GetBestHost()
	if err != nil {
		log.Println(err)
		return ""
	}
	u, err := url.Parse(host + path)
	if nil != err {
		log.Println(err)
		return ""
	}
	return u.String()
}
