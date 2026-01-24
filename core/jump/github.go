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
func ParseGithubProxyPath(proxyPath string) (string, string, error) {
	parts := strings.SplitN(proxyPath, "/", 2)
	if len(parts) < 2 {
		return "", "", errors.New("invalid proxy path")
	}

	targetHost := parts[0]
	targetPath := "/" + parts[1]

	// Host 白名单校验
	if !githubAllowHosts[targetHost] {
		return "", "", errors.New("host not allowed")
	}

	return targetHost, targetPath, nil
}
