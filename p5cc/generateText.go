package p5cc

import (
	"image/color"
	"math"
	"math/rand"

	"github.com/fogleman/gg"
)

// 颜色常量
var (
	ColorRed   = color.RGBA{R: 0xE5, G: 0x19, B: 0x1C, A: 0xFF}
	ColorWhite = color.RGBA{R: 0xFD, G: 0xFD, B: 0xFD, A: 0xFF}
	ColorBlack = color.RGBA{R: 0x0F, G: 0x0F, B: 0x0F, A: 0xFF}
)

const (
	CharModeFirst = iota + 1
	CharModeWhite
	CharModeRed
	CharModeSpace
)

type BoxChar struct {
	Char       string
	FontSize   float64
	FontFamily string
	Width      float64
	Height     float64
	Angle      float64
	Scale      float64
	Mode       int
	Color      color.Color
}

type BoxText struct {
	Chars      []*BoxChar
	FontSize   float64
	FontFamily string
	Gutter     float64
	Padding    float64
	TextAlign  string
}

func NewBoxChar(char string, mode int, fontSize float64, fontFamily string) *BoxChar {
	bc := &BoxChar{
		Char:       char,
		Mode:       mode,
		FontFamily: fontFamily,
	}

	if mode == CharModeSpace {
		return bc
	}

	// 随机旋转和缩放
	angle := -10 + rand.Float64()*20 // [-10°, 10°]
	if mode == CharModeFirst {
		bc.Scale = 1.1
		bc.Angle = angle
	} else {
		bc.Scale = 1.0 - rand.Float64()*0.3 // 0.7~1.0
		bc.Angle = angle * randOper()
	}

	bc.FontSize = fontSize * bc.Scale
	bc.Color = ColorWhite
	if mode == CharModeRed {
		bc.Color = ColorRed
	}

	// 测量字符尺寸
	dc := gg.NewContext(1, 1)
	dc.LoadFontFace(bc.FontFamily, bc.FontSize)
	w, h := dc.MeasureString(bc.Char)
	bc.Width, bc.Height = w, h

	return bc
}

func randOper() float64 {
	if rand.Float32() < 0.5 {
		return 1
	}
	return -1
}

func (bc *BoxChar) OutterSize() (float64, float64) {
	rad := bc.Angle * math.Pi / 180
	sin := math.Abs(math.Sin(rad))
	cos := math.Abs(math.Cos(rad))
	w := bc.Width*cos + bc.Height*sin
	h := bc.Height*cos + bc.Width*sin

	scale := 1.2
	if bc.Mode == CharModeFirst {
		scale = 1.4
	}
	return w * scale, h * scale
}

func NewBoxText(text string, options map[string]interface{}) *BoxText {
	bt := new(BoxText)
	// 应用配置
	for k, v := range options {
		switch k {
		case "fontSize":
			bt.FontSize = v.(float64)
		case "fontFamily":
			bt.FontFamily = v.(string)
		case "gutter":
			bt.Gutter = v.(float64)
		case "padding":
			bt.Padding = v.(float64)
		case "textAlign":
			bt.TextAlign = v.(string)
		}
	}

	// 初始化字符模式
	chars := []rune(text)
	modes := make([]int, len(chars))
	modes[0] = CharModeFirst

	// 随机设置红色字符
	for i := 1; i < len(chars); i++ {
		if rand.Float32() < 0.33 {
			modes[i] = CharModeRed
		} else {
			modes[i] = CharModeWhite
		}
	}

	// 创建 BoxChar 实例
	for i, c := range chars {
		char := string(c)
		if char == " " {
			bt.Chars = append(bt.Chars, NewBoxChar("", CharModeSpace, 0, ""))
		} else {
			bc := NewBoxChar(char, modes[i], bt.FontSize, bt.FontFamily)
			bt.Chars = append(bt.Chars, bc)
		}
	}

	return bt
}

func (bt *BoxText) Draw(dc *gg.Context) float64 {
	padding := bt.Padding
	gutter := bt.Gutter
	totalWidth := 2 * padding
	maxHeight := 0.0
	isNewLine := false // 是否换行

	// 计算布局
	for _, bc := range bt.Chars {
		if bc.Mode == CharModeSpace {
			totalWidth += 2 * gutter
			continue
		}
		w, h := bc.OutterSize()
		totalWidth += w + gutter
		// 判断是否换行
		if totalWidth > float64(dc.Width())-2*padding {
			totalWidth = 2 * padding
			maxHeight += h + 2*gutter
			isNewLine = true
		}
	}

	// 对齐计算
	startX := padding
	switch bt.TextAlign {
	case "center":
		if isNewLine {
			startX = padding
		} else {
			startX = (float64(dc.Width()) - totalWidth + 2*padding) / 2
		}
	case "left":
		startX = padding
	}

	// 绘制字符
	x := startX
	y := float64(dc.Height())/2
	maxH := 0.0 // 第一个字符的高度
	for _, bc := range bt.Chars {
		if bc.Mode == CharModeSpace {
			x += 2 * gutter
			continue
		}

		outterW, outterH := bc.OutterSize()
		if outterH > maxH {
			maxH = outterH
			y = y - maxH
		}

		// 判断是否换行
		if x+outterW > float64(dc.Width())-padding {
			x = startX
			y += maxH + 2*gutter
		}

		// 绘制白色边框
		dc.Push()
		dc.RotateAbout(gg.Radians(bc.Angle), x+outterW/2, y+outterH/2)
		dc.SetColor(ColorWhite)
		dc.DrawRectangle(x, y, outterW, outterH)
		dc.Fill()
		// 绘制黑色边框
		dc.Pop()
		dc.Push()
		dc.RotateAbout(gg.Radians(bc.Angle), x+outterW/2, y+outterH/2)
		dc.SetColor(ColorBlack)
		dc.DrawRectangle(x+5, y+5, outterW-10, outterH-10)
		dc.Fill()
		dc.Pop()
		if bc.Mode == CharModeFirst {
			dc.Push()
			dc.RotateAbout(gg.Radians(bc.Angle), x+outterW/2, y+outterH/2)
			dc.SetColor(ColorRed)
			dc.DrawRectangle(x+10, y+10, outterW-20, outterH-20)
			dc.Fill()
			dc.Pop()
		}

		// 绘制文字
		dc.Push()
		dc.RotateAbout(gg.Radians(bc.Angle), x+outterW/2, y+outterH/2)
		dc.SetColor(bc.Color)
		if bc.Mode == CharModeFirst {
			dc.LoadFontFace(bt.FontFamily, bc.FontSize)
		} else {
			dc.LoadFontFace(bt.FontFamily, bc.FontSize*0.85)
		}
		dc.DrawStringAnchored(bc.Char, x+outterW/2, y+outterH/2, 0.5, 0.5)
		dc.Pop()

		x += outterW + gutter
	}

	return maxHeight + 2*padding
}
