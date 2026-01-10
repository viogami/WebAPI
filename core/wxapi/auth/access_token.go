package auth

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"time"
)

var Access_Token string //access_token全局变量

// AccessTokenResponse 用于解析获取 Access Token 的响应
type AccessTokenResponse struct {
	AccessToken string `json:"access_token"`
	ExpiresIn   int    `json:"expires_in"`
}

// GetAccessToken 获取微信公众号的 Access Token
func GetAccessToken(appID, appSecret string) (string, int, error) {
	// 构建请求 URL
	url := fmt.Sprintf("https://api.weixin.qq.com/cgi-bin/token?grant_type=client_credential&appid=%s&secret=%s", appID, appSecret)

	// 发送 HTTP GET 请求
	resp, err := http.Get(url)
	if err != nil {
		return "", 0, err
	}
	defer resp.Body.Close()

	// 读取响应体
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", 0, err
	}

	// 解析 JSON 响应
	var accessTokenResponse AccessTokenResponse
	err = json.Unmarshal(body, &accessTokenResponse)
	if err != nil {
		return "", 0, err
	}

	return accessTokenResponse.AccessToken, accessTokenResponse.ExpiresIn, err
}

// 定时获取微信 Access Token
func StartAccessTokenScheduler(appID, appSecret string) {
	// 启动一个定时任务，每隔 2 小时执行一次获取 Access Token 的操作
	ticker := time.NewTicker(2 * time.Hour)
	defer ticker.Stop() // 停止定时器

	for range ticker.C {
		// 调用 GetAccessToken 获取 Access Token
		accessToken, expiresIn, err := GetAccessToken(appID, appSecret)
		if err != nil {
			log.Println("Error getting Access Token:", err)
			continue
		}
		Access_Token = accessToken // 更新 accessToken
		log.Println("Now the Access Token:", accessToken, " ExpiresIn:", expiresIn)
	}
}
