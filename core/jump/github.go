package jump

import (
	"errors"
	"strings"
)

// GitHub 允许代理的 Host 白名单（强烈建议）
var githubAllowHosts = map[string]bool{
	"github.com":                     true,
	"raw.githubusercontent.com":      true,
	"github-readme-stats.vercel.app": true,
}

// /jump/github/{host}/{path...} github 代理路径解析
func BuildGithubURL(proxyPath, rawQuery string) (string, error) {
	parts := strings.SplitN(proxyPath, "/", 2)
	if len(parts) < 2 {
		return "", errors.New("invalid proxy path")
	}

	host := parts[0]
	path := parts[1]
	
	if !githubAllowHosts[host] {
		return "", errors.New("host not allowed")
	}

	u := "https://" + host + "/" + path
	if rawQuery != "" {
		u += "?" + rawQuery
	}
	return u, nil
}
