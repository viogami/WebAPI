package handlers

import (
	"WebAPI/conf"
	"WebAPI/core/wxapi"
	"WebAPI/core/wxapi/auth"
	"log"

	"github.com/gin-gonic/gin"
	"github.com/silenceper/wechat/v2"
	m "github.com/silenceper/wechat/v2/officialaccount/menu"
)

type WxapiHandler struct {
	wx *wechat.Wechat
}

func (w *WxapiHandler) Hello(c *gin.Context) {
	c.JSON(200, conf.AppConfig.Wxapi.Text.HelloText)
}

func (w *WxapiHandler) Redirect(c *gin.Context) {
	c.Redirect(302, "/")
}

// GetAccessToken 获取ak
func (w *WxapiHandler) GetAccessToken(c *gin.Context) {
	oa := w.wx.GetOfficialAccount(&conf.AppConfig.Wxapi.OfficialAccount)
	token, err := oa.GetAccessToken()
	if err != nil {
		c.JSON(500, gin.H{"error": err.Error()})
		return
	}
	c.JSON(200, gin.H{"access_token": token})
}

// GetCallbackIP 获取微信callback IP地址
func (w *WxapiHandler) GetCallbackIP(c *gin.Context) {
	oa := w.wx.GetOfficialAccount(&conf.AppConfig.Wxapi.OfficialAccount)
	ips, err := oa.GetBasic().GetCallbackIP()
	if err != nil {
		c.JSON(500, gin.H{"error": err.Error()})
		return
	}
	c.JSON(200, gin.H{"ip_list": ips})
}

// GetAPIDomainIP 获取微信callback IP地址
func (w *WxapiHandler) GetAPIDomainIP(c *gin.Context) {
	oa := w.wx.GetOfficialAccount(&conf.AppConfig.Wxapi.OfficialAccount)
	ips, err := oa.GetBasic().GetAPIDomainIP()
	if err != nil {
		c.JSON(500, gin.H{"error": err.Error()})
		return
	}
	c.JSON(200, gin.H{"ip_list": ips})
}

// GetAPIDomainIP  清理接口调用次数
func (w *WxapiHandler) ClearQuota(c *gin.Context) {
	oa := w.wx.GetOfficialAccount(&conf.AppConfig.Wxapi.OfficialAccount)
	err := oa.GetBasic().ClearQuota()
	if err != nil {
		c.JSON(500, gin.H{"error": err.Error()})
		return
	}
	c.JSON(200, gin.H{"message": "clear quota success"})
}

// createMenu 创建菜单
func (w *WxapiHandler) createMenu(c *gin.Context) {
	buttons := make([]*m.Button, 0)
	// 一级菜单
	btn1 := m.NewClickButton("bot信息", "vio--made by viogami")
	btn2 := m.NewViewButton("作者博客", "http://viogami.tech/")
	buttons = append(buttons, btn1, btn2)

	oa := w.wx.GetOfficialAccount(&conf.AppConfig.Wxapi.OfficialAccount)
	err := oa.GetMenu().SetMenu(buttons)
	if err != nil {
		c.JSON(500, "菜单创建失败"+err.Error())
		return
	}
	c.JSON(200, "菜单创建成功")
}

// CheckMenu 检查菜单
func (w *WxapiHandler) CheckMenu(c *gin.Context) {
	oa := w.wx.GetOfficialAccount(&conf.AppConfig.Wxapi.OfficialAccount)
	menu, err := oa.GetMenu().GetMenu()
	if err != nil {
		// 创建菜单
		w.createMenu(c)
		return
	}
	c.JSON(200, menu)
}

// DeleteMenu 删除菜单
func (w *WxapiHandler) DeleteMenu() error {
	oa := w.wx.GetOfficialAccount(&conf.AppConfig.Wxapi.OfficialAccount)
	oa.GetMenu().DeleteMenu()
	return nil
}

// AddConditionalMenu 添加个性化菜单
func (w *WxapiHandler) AddConditionalMenu(buttons []*m.Button, matchRule *m.MatchRule) error {
	oa := w.wx.GetOfficialAccount(&conf.AppConfig.Wxapi.OfficialAccount)
	oa.GetMenu().AddConditional(buttons, matchRule)
	return nil
}

// 微信接入校验
func (w *WxapiHandler) WXCheckSignature(c *gin.Context) {
	signature := c.Query("signature")
	timestamp := c.Query("timestamp")
	nonce := c.Query("nonce")
	echostr := c.Query("echostr")
	wxToken := conf.AppConfig.Wxapi.OfficialAccount.Token

	ok := auth.CheckSignature(signature, timestamp, nonce, wxToken)
	if !ok {
		log.Println("微信公众号签名检查失败!")
		return
	}

	log.Println("微信公众号接入校验成功!")
	_, _ = c.Writer.WriteString(echostr)
}

// WXMsgReceive 微信消息接收
func (w *WxapiHandler) WXMsgReceive(c *gin.Context) {
	var textMsg *wxapi.WxTextMsg
	err := c.ShouldBindXML(&textMsg)
	if err != nil {
		log.Printf("[消息接收] - XML数据包解析失败: %v\n", err)
		return
	}
	log.Printf("[消息接收] - 收到消息, 消息类型为: %s, 消息内容为: %s\n", textMsg.MsgType, textMsg.Content)

	msg := textMsg.WXMsgReceive()
	_, _ = c.Writer.Write(msg)
}

func NewWxapiHandler() *WxapiHandler {
	return &WxapiHandler{
		wx: wechat.NewWechat(),
	}
}
