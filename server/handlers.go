package server

import (
	"WebAPI/p5cc"
	"image/png"
	"net/http"

	"github.com/gin-gonic/gin"
)

func helloHandler(c *gin.Context) {
	c.String(http.StatusOK, AppConfig.TextConfig.HelloText)
}

// p5cc Get请求
func p5ccHandler(c *gin.Context) {
	cfg := AppConfig.P5cc
	p5cfg := p5cc.P5ccConfig{
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
	img, err := p5cc.GenerateCard(text, p5cfg)
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
func UpdateP5ccHandler(c *gin.Context) {

}
