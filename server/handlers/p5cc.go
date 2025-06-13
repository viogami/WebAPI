package handlers

import (
	"WebAPI/conf"
	"WebAPI/core/p5cc"
	"image/png"
	"net/http"

	"github.com/gin-gonic/gin"
)

type P5ccHandler struct {
	
}

// p5cc Get请求
func (p *P5ccHandler) GET(c *gin.Context) {
	cfg := conf.AppConfig.P5cc
	option := p5cc.P5ccConfig{
		FontSize:   cfg.FontSize,
		FontFamily: cfg.FontFamily,
		Gutter:     cfg.Gutter,
		Padding:    cfg.Padding,
		TextAlign:  cfg.TextAlign,
		RedProb:    cfg.RedProb,
		ShowLogo:   cfg.ShowLogo,
		LogoScale:  cfg.LogoScale,
		LogoOffset: cfg.LogoOffset,
		ShowWtm:    cfg.ShowWtm,
	}
	text := "TAKEYOUHEART"
	if c.Param("text") != "" {
		text = c.Param("text")
	}

	// 生成卡片
	img, err := p5cc.GenerateCard(text, option)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})
		return
	}

	// 返回 PNG 图像
	c.Writer.Header().Set("Content-Type", "image/png")
	png.Encode(c.Writer, img)
}

// p5cc Post请求
func (p *P5ccHandler) POST(c *gin.Context) {
	// 在路由处理函数中使用
	defaultCfg := conf.AppConfig.P5cc
	text := c.PostForm("text")

	option := p5cc.P5ccConfig{
		FontSize:   getFloatParam(c, "fontSize", defaultCfg.FontSize),
		FontFamily: defaultCfg.FontFamily, // 不支持修改字体
		Gutter:     getFloatParam(c, "gutter", defaultCfg.Gutter),
		Padding:    getFloatParam(c, "padding", defaultCfg.Padding),
		TextAlign:  c.DefaultPostForm("textAlign", defaultCfg.TextAlign),
		RedProb:    getFloatParam(c, "redProb", defaultCfg.RedProb),
		ShowLogo:   getBoolParam(c, "showLogo", defaultCfg.ShowLogo),
		LogoScale:  getFloatParam(c, "logoScale", defaultCfg.LogoScale),
		LogoOffset: getIntParam(c, "logoOffset", defaultCfg.LogoOffset),
		ShowWtm:    c.DefaultPostForm("showWtm", defaultCfg.ShowWtm),
	}
	if text == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "text is required",
		})
		return
	}

	// 生成卡片
	img, err := p5cc.GenerateCard(text, option)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})
		return
	}
	// 返回 PNG 图像
	c.Writer.Header().Set("Content-Type", "image/png")
	png.Encode(c.Writer, img)

}

func NewP5ccHandler() *P5ccHandler {
	return &P5ccHandler{}
}