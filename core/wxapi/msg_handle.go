package wxapi

import (
	"WebAPI/conf"
	"encoding/xml"
	"io"
	"log"
	"net/http"
	"net/url"
	"time"
)

// WxTextMsg 微信文本消息结构体
type WxTextMsg struct {
	ToUserName   string
	FromUserName string
	CreateTime   int64
	MsgType      string
	Content      string
	MsgId        int64
}

// WXMsgReceive 微信消息接收
func (textMsg *WxTextMsg) WXMsgReceive() []byte {
	switch textMsg.MsgType {
	case "event":
		return simpleReply(textMsg.ToUserName, textMsg.FromUserName, conf.AppConfig.Wxapi.Text.SimplayText) // 关注后的默认回复

	case "text":
		return gptReplyWXMsg(textMsg.ToUserName, textMsg.FromUserName, textMsg.Content) // 调用gpt回复

	default:
		return simpleReply(textMsg.ToUserName, textMsg.FromUserName, conf.AppConfig.Wxapi.Text.DefaultText) // 未定义类型的回复
	}
}

// WXRepTextMsg 微信回复文本消息结构体
type WXRepTextMsg struct {
	ToUserName   string
	FromUserName string
	CreateTime   int64
	MsgType      string
	Content      string
	// 若不标记XMLName, 则解析后的xml名为该结构体的名称
	XMLName xml.Name `xml:"xml"`
}

// gptReplyWXMsg 微信消息回复
func gptReplyWXMsg(fromUser, toUser, usercontent string) []byte {
	//定义post的url地址
	URL := conf.AppConfig.Wxapi.ReplyPostURL
	// 准备POST请求的数据
	formData := url.Values{
		"usermsg": {usercontent},
	}
	// 发送POST请求
	resp, err := http.PostForm(URL, formData)
	if err != nil {
		log.Println("Error sending POST request:", err)
		return nil
	}
	defer resp.Body.Close()
	respbody, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Println("Error to read the response body", err)
		return nil
	}

	repTextMsg := WXRepTextMsg{
		ToUserName:   toUser,
		FromUserName: fromUser,
		CreateTime:   time.Now().Unix(),
		MsgType:      "text",
		Content:      string(respbody),
	}

	msg, err := xml.Marshal(&repTextMsg)
	if err != nil {
		log.Printf("[ERROR] - 将对象进行XML编码出错: %v\n", err)
		return nil
	}
	return msg
}

func simpleReply(fromUser, toUser, text string) []byte {
	repTextMsg := WXRepTextMsg{
		ToUserName:   toUser,
		FromUserName: fromUser,
		CreateTime:   time.Now().Unix(),
		MsgType:      "text",
		Content:      text,
	}

	msg, err := xml.Marshal(&repTextMsg)
	if err != nil {
		log.Printf("[ERROR] - 将对象进行XML编码出错: %v\n", err)
		return nil
	}
	return msg
}
