package provider

import (
	"net/http"
	"net/http/httptrace"
	"net/url"
	"time"

	"golang.org/x/net/html"
)

func getAttrFromNode(node *html.Node, attr string) string {
	for _, a := range node.Attr {
		if a.Key == attr {
			return a.Val
		}
	}
	return ""
}

func pingHost(uri *url.URL) (int, error) {
	req, _ := http.NewRequest("GET", uri.String(), nil)
	var start time.Time
	trace := &httptrace.ClientTrace{}
	req = req.WithContext(httptrace.WithClientTrace(req.Context(), trace))
	start = time.Now()
	if _, err := http.DefaultTransport.RoundTrip(req); err != nil {
		return 0, err
	}
	return int(time.Since(start).Milliseconds()), nil
}
