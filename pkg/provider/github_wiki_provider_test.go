package provider

import (
	"fmt"
	"testing"
)

func TestGithubWikiProvider(t *testing.T) {
	githubProvider := NewGithubProvider()
	cfg := map[string]interface{}{
		"expr": `//*[@id="wiki-body"]/div[1]/table[2]/tbody/tr[*]`,
	}
	err := githubProvider.Init(cfg)
	if err != nil {
		t.Error(err)
	}
	activeHosts, err := githubProvider.GetActiveHosts()
	if err != nil {
		t.Error(err)
	}
	fmt.Println(activeHosts)
	baseHost, err := githubProvider.GetBestHost()
	if err != nil {
		t.Error(err)
	}
	fmt.Println(baseHost)
	oneHost, err := githubProvider.RandomHost()
	if err != nil {
		t.Error(err)
	}
	fmt.Println(oneHost)

}
