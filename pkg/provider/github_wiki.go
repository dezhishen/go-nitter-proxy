package provider

import (
	"fmt"
	"math/rand"
	"net/http"
	"net/url"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"github.com/antchfx/htmlquery"
	"golang.org/x/net/html"
	"golang.org/x/net/html/charset"
)

type hostWithStatus struct {
	*url.URL
	// 状态
	status bool
	// 延迟 ms
	delay int
}
type githubProvider struct {
	hosts []*hostWithStatus
	// 仓库地址
	repo string
	// 仓库代理
	repoProxy *url.URL
	// 仓库http客户端
	repoHttpClient *http.Client
	// 解析表达式
	expr string
}

func NewGithubProvider() HostProvider {
	return &githubProvider{
		repo: "https://github.com/zedeus/nitter",
		expr: `//*[@id="wiki-body"]/div[1]/table[2]/tbody/tr[*]`,
	}
}

func (p *githubProvider) Init(cfg map[string]interface{}) error {
	if cfg != nil {
		if repo, ok := cfg["repo"]; ok {
			p.repo = repo.(string)
		}
		if expr, ok := cfg["expr"]; ok {
			p.expr = expr.(string)
		}
		if repoProxy, ok := cfg["repoProxy"]; ok {
			uri, err := url.Parse(repoProxy.(string))
			if err != nil {
				fmt.Println("repo proxy is err", err)
			} else {
				p.repoProxy = uri
			}
		}
	}
	if p.repoProxy == nil {
		p.repoHttpClient = http.DefaultClient
	} else {
		p.repoHttpClient = &http.Client{
			Transport: &http.Transport{
				// 设置代理，从环境变量中获取
				Proxy: http.ProxyURL(p.repoProxy),
			},
		}

	}
	// gethosts from repo
	resetHostOnce(p)
	go p.testHosts()
	return nil
}

func resetHostOnce(p *githubProvider) {
	repoUrl := p.repo + "/wiki/instances"
	resp, err := p.repoHttpClient.Get(repoUrl)
	if err != nil {
		fmt.Println("open hosts web error", err)
		return
	}
	defer resp.Body.Close()
	r, err := charset.NewReader(resp.Body, resp.Header.Get("Content-Type"))
	if err != nil {
		return
	}
	doc, err := html.Parse(r)
	if err != nil {
		fmt.Println("parse hosts web err", err)
		return
	}
	xpath := p.expr
	nodes, err := htmlquery.QueryAll(doc, xpath)
	if err != nil {
		fmt.Println("get hosts list error ", err)
		return
	}
	if len(nodes) == 0 {
		fmt.Println("hosts list is empty")
		return
	}
	node2host := func(node *html.Node) *hostWithStatus {
		//alias="white_check_mark"
		statusNode := htmlquery.FindOne(node, "//td[2]")
		if statusNode == nil {
			return nil
		}
		statusString := getAttrFromNode(statusNode.FirstChild, "alias")
		status := (statusString == "white_check_mark")
		urlNode := htmlquery.FindOne(node, "//td[1]")
		urlString := getAttrFromNode(urlNode.FirstChild, "href")
		if urlString == "" {
			return nil
		}
		uri, err := url.Parse(urlString)
		if err != nil {
			return nil
		}
		delay, err := pingHost(uri)
		if err != nil {
			status = false
		}

		return &hostWithStatus{
			URL:    uri,
			status: status,
			delay:  delay,
		}
	}
	var result []*hostWithStatus
	rw := sync.RWMutex{}
	wg := sync.WaitGroup{}
	wg.Add(len(nodes))
	// 并发获取延迟,最多同时运行10个
	ch := make(chan bool, 30)
	nodeLen := len(nodes)
	for i, node := range nodes {
		ch <- true
		index := i + 1
		node := node
		go func() {
			defer func() {
				wg.Done()
				<-ch
			}()
			fmt.Printf("start [%d/%d]\n", index, nodeLen)
			fmt.Println("start parse node ", index, "/", nodeLen)
			e := node2host(node)
			if e == nil {
				return
			}
			fmt.Printf("succes [%d/%d] url:[%s] status:[%v] deloy[%d]\n",
				index,
				nodeLen,
				e.URL.String(),
				e.status,
				e.delay,
			)
			rw.Lock()
			defer rw.Unlock()
			result = append(result, e)
		}()

	}
	wg.Wait()
	fmt.Printf("reset success len: %d", len(result))
	rw.RLock()
	defer rw.RUnlock()
	p.hosts = result
}

func resetHosts(p *githubProvider) {
	for {
		time.Sleep(30 * time.Minute)
		resetHostOnce(p)
	}
}

func pingHosts(p *githubProvider) {
	for {
		time.Sleep(5 * time.Minute)
		for _, hws := range p.hosts {
			deloy, err := pingHost(hws.URL)
			if err != nil {
				hws.status = false
			} else {
				hws.status = true
				hws.delay = deloy
			}

		}
	}
}

func (p *githubProvider) testHosts() {
	go resetHosts(p)
	go pingHosts(p)
	c := make(chan os.Signal, 1)
	//监听指定信号 ctrl+c kill
	signal.Notify(c, os.Interrupt, syscall.SIGHUP, syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT)
	//阻塞直至有信号传入
	<-c
}

func (p *githubProvider) GetActiveHosts() ([]string, error) {
	var result []string
	for _, hws := range p.hosts {
		if hws.status {
			result = append(result, hws.String())
		}
	}
	return result, nil
}

func (p *githubProvider) GetBestHost() (string, error) {
	var best *hostWithStatus
	for _, hws := range p.hosts {
		if hws.status {
			if best == nil {
				best = hws
			} else {
				if hws.delay < best.delay {
					best = hws
				}
			}
		}
	}
	if best == nil {
		return "", nil
	}
	return best.String(), nil

}

func (p *githubProvider) RandomHost() (string, error) {
	activeHost, err := p.GetActiveHosts()
	if err != nil {
		return "", err
	}
	if len(activeHost) == 0 {
		return "", nil
	}
	return activeHost[rand.Intn(len(activeHost))], nil
}
