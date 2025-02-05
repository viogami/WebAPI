package server

import (
	"WebAPI/p5cc"
	"image/png"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

func helloHandler(c *gin.Context) {
	c.String(http.StatusOK, AppConfig.TextConfig.HelloText)
}

// p5cc Get请求
func p5ccHandler(c *gin.Context) {
	cfg := AppConfig.P5cc
	option := p5cc.P5ccConfig{
		FontSize:   cfg.FontSize,
		FontFamily: cfg.FontFamily,
		Gutter:     cfg.Gutter,
		Padding:    cfg.Padding,
		TextAlign:  cfg.TextAlign,
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
func P5ccPostHandler(c *gin.Context) {
	// 在路由处理函数中使用
	defaultCfg := AppConfig.P5cc
	text := c.PostForm("text")

	option := p5cc.P5ccConfig{
		FontSize:   getFloatParam(c, "fontSize", defaultCfg.FontSize),
		FontFamily: defaultCfg.FontFamily, // 不支持修改字体
		Gutter:     getFloatParam(c, "gutter", defaultCfg.Gutter),
		Padding:    getFloatParam(c, "padding", defaultCfg.Padding),
		TextAlign:  c.DefaultPostForm("textAlign", defaultCfg.TextAlign),
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

// 创建辅助函数处理不同类型参数获取
func getFloatParam(c *gin.Context, param string, defaultValue float64) float64 {
	if paramStr := c.PostForm(param); paramStr != "" {
		if val, err := strconv.ParseFloat(paramStr, 64); err == nil {
			return val
		}
	}
	return defaultValue
}

func getBoolParam(c *gin.Context, param string, defaultValue bool) bool {
	if paramStr := c.PostForm(param); paramStr != "" {
		if val, err := strconv.ParseBool(paramStr); err == nil {
			return val
		}
	}
	return defaultValue
}

func getIntParam(c *gin.Context, param string, defaultValue int) int {
	if paramStr := c.PostForm(param); paramStr != "" {
		if val, err := strconv.Atoi(paramStr); err == nil {
			return val
		}
	}
	return defaultValue
}
