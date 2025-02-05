package p5cc

import (
	"fmt"
	"image"
	"image/color"
	"os"

	"github.com/fogleman/gg"
)

type P5ccConfig struct {
	FontSize   float64 
	FontFamily string  
	Gutter     float64 
	Padding    float64 
	TextAlign  string  

	ShowLogo   bool    
	LogoScale  float64 
	LogoOffset int     

	ShowWtm string 
}

// 加载本地图像文件
func loadImage(path string) (image.Image, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	img, _, err := image.Decode(file)
	if err != nil {
		return nil, err
	}
	return img, nil
}

// 生成卡片
func GenerateCard(text string, option P5ccConfig) (image.Image, error) {
	const (
		canvasWidth  = 2200
		canvasHeight = 1672
	)

	// 创建画布
	dc := gg.NewContext(canvasWidth, canvasHeight)
	// 加载画布
	canvas ,err := loadImage("p5cc/assets/canvas.png")
	if err != nil {
		return nil, fmt.Errorf("failed to load canvas image: %v", err)
	}
	canvasX := (canvasWidth - canvas.Bounds().Dx()) / 2
	canvasY := (canvasHeight - canvas.Bounds().Dy()) / 2
	dc.DrawImage(canvas, canvasX, canvasY)

	// 加载背景图像
	baseCard, err := loadImage("p5cc/assets/base.png")
	if err != nil {
		return nil, fmt.Errorf("failed to load base image: %v", err)
	}
	baseX := (canvasWidth - baseCard.Bounds().Dx()) / 2
	baseY := (canvasHeight - baseCard.Bounds().Dy()) / 2
	dc.DrawImage(baseCard, baseX, baseY)

	// 加载并绘制 Logo
	if option.ShowLogo {
		logo, err := loadImage("p5cc/assets/logo.png")
		if err != nil {
			return nil, fmt.Errorf("failed to load logo image: %v", err)
		}
		logoWidth := float64(logo.Bounds().Dx()) * option.LogoScale
		logoHeight := float64(logo.Bounds().Dy()) * option.LogoScale
		dc.DrawImageAnchored(logo, canvasWidth-int(logoWidth)-int(option.LogoOffset), canvasHeight-int(logoHeight)-int(option.LogoOffset), 0, 0)
	}

	// 绘制水印文本
	if option.ShowWtm != "" {
		dc.SetColor(color.RGBA{255, 255, 255, 165}) // 半透明白色
		dc.LoadFontFace(option.FontFamily, 24)
		dc.DrawStringAnchored(option.ShowWtm, 30, float64(canvasHeight)-30, 0, 0)
	}

	// 绘制用户文本
	options := map[string]interface{}{
		"fontSize":   option.FontSize,
		"fontFamily": option.FontFamily,
		"gutter":     option.Gutter,
		"padding":    option.Padding,
		"textAlign":  option.TextAlign,
	}
	boxText := NewBoxText(text, options)
	boxText.Draw(dc)

	return dc.Image(), nil
}
