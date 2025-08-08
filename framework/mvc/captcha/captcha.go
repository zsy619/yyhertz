package captcha

import (
	"bytes"
	"crypto/rand"
	"fmt"
	"image"
	"image/color"
	"image/draw"
	"image/png"
	"math"
	"math/big"
	"strings"
	"time"
)

// Config 验证码配置
type Config struct {
	Width   int    // 图片宽度
	Height  int    // 图片高度
	Length  int    // 验证码长度
	TTL     int64  // 过期时间(秒)
	Charset string // 字符集
}

// DefaultConfig 默认配置
func DefaultConfig() *Config {
	return &Config{
		Width:   120,
		Height:  40,
		Length:  4,
		TTL:     300, // 5分钟
		Charset: "0123456789",
	}
}

// Captcha 验证码实例
type Captcha struct {
	ID     string
	Code   string
	Image  []byte
	ExpireAt int64
}

// Generator 验证码生成器
type Generator struct {
	config *Config
	store  Store
}

// Store 验证码存储接口
type Store interface {
	Set(id string, captcha *Captcha) error
	Get(id string) (*Captcha, error)
	Delete(id string) error
	Clear() error
}

// NewGenerator 创建验证码生成器
func NewGenerator(config *Config, store Store) *Generator {
	if config == nil {
		config = DefaultConfig()
	}
	return &Generator{
		config: config,
		store:  store,
	}
}

// Generate 生成验证码
func (g *Generator) Generate() (*Captcha, error) {
	id := generateID()
	code := g.generateCode()
	imageBytes, err := g.generateImage(code)
	if err != nil {
		return nil, err
	}
	
	captcha := &Captcha{
		ID:       id,
		Code:     code,
		Image:    imageBytes,
		ExpireAt: time.Now().Unix() + g.config.TTL,
	}
	
	if err := g.store.Set(id, captcha); err != nil {
		return nil, err
	}
	
	return captcha, nil
}

// Verify 验证验证码
func (g *Generator) Verify(id, code string) bool {
	captcha, err := g.store.Get(id)
	if err != nil {
		return false
	}
	
	// 检查过期时间
	if time.Now().Unix() > captcha.ExpireAt {
		g.store.Delete(id)
		return false
	}
	
	// 验证成功后删除验证码
	defer g.store.Delete(id)
	
	// 不区分大小写
	return strings.EqualFold(captcha.Code, code)
}

// GetImage 获取验证码图片
func (g *Generator) GetImage(id string) ([]byte, error) {
	captcha, err := g.store.Get(id)
	if err != nil {
		return nil, err
	}
	
	// 检查过期时间
	if time.Now().Unix() > captcha.ExpireAt {
		g.store.Delete(id)
		return nil, fmt.Errorf("captcha expired")
	}
	
	return captcha.Image, nil
}

// generateID 生成唯一ID
func generateID() string {
	return fmt.Sprintf("%d", time.Now().UnixNano())
}

// generateCode 生成验证码字符串
func (g *Generator) generateCode() string {
	chars := []rune(g.config.Charset)
	code := make([]rune, g.config.Length)
	
	for i := 0; i < g.config.Length; i++ {
		n, _ := rand.Int(rand.Reader, big.NewInt(int64(len(chars))))
		code[i] = chars[n.Int64()]
	}
	
	return string(code)
}

// generateImage 生成验证码图片
func (g *Generator) generateImage(code string) ([]byte, error) {
	img := image.NewRGBA(image.Rect(0, 0, g.config.Width, g.config.Height))
	
	// 填充背景
	draw.Draw(img, img.Bounds(), &image.Uniform{color.RGBA{240, 240, 240, 255}}, image.Point{}, draw.Src)
	
	// 添加噪点
	g.addNoise(img)
	
	// 绘制文字
	g.drawText(img, code)
	
	// 添加干扰线
	g.addLines(img)
	
	// 转换为PNG字节数组
	var buf bytes.Buffer
	if err := png.Encode(&buf, img); err != nil {
		return nil, err
	}
	
	return buf.Bytes(), nil
}

// addNoise 添加噪点
func (g *Generator) addNoise(img *image.RGBA) {
	bounds := img.Bounds()
	for i := 0; i < 100; i++ {
		x, _ := rand.Int(rand.Reader, big.NewInt(int64(bounds.Max.X)))
		y, _ := rand.Int(rand.Reader, big.NewInt(int64(bounds.Max.Y)))
		
		r, _ := rand.Int(rand.Reader, big.NewInt(256))
		g, _ := rand.Int(rand.Reader, big.NewInt(256))
		b, _ := rand.Int(rand.Reader, big.NewInt(256))
		
		img.Set(int(x.Int64()), int(y.Int64()), color.RGBA{
			uint8(r.Int64()), uint8(g.Int64()), uint8(b.Int64()), 255,
		})
	}
}

// drawText 绘制文字
func (g *Generator) drawText(img *image.RGBA, text string) {
	bounds := img.Bounds()
	charWidth := bounds.Max.X / len(text)
	
	colors := []color.RGBA{
		{255, 0, 0, 255},   // 红色
		{0, 255, 0, 255},   // 绿色
		{0, 0, 255, 255},   // 蓝色
		{255, 0, 255, 255}, // 紫色
		{255, 165, 0, 255}, // 橙色
	}
	
	for i, char := range text {
		x := i*charWidth + charWidth/4
		y := bounds.Max.Y/2 - 5
		
		// 随机颜色
		colorIndex, _ := rand.Int(rand.Reader, big.NewInt(int64(len(colors))))
		c := colors[colorIndex.Int64()]
		
		// 简单的点阵字体绘制
		g.drawChar(img, char, x, y, c)
	}
}

// drawChar 绘制单个字符
func (g *Generator) drawChar(img *image.RGBA, char rune, x, y int, c color.RGBA) {
	// 简单的数字点阵数据（这里只实现0-9）
	patterns := map[rune][][]int{
		'0': {
			{0, 1, 1, 0},
			{1, 0, 0, 1},
			{1, 0, 0, 1},
			{1, 0, 0, 1},
			{0, 1, 1, 0},
		},
		'1': {
			{0, 1, 0, 0},
			{1, 1, 0, 0},
			{0, 1, 0, 0},
			{0, 1, 0, 0},
			{1, 1, 1, 0},
		},
		'2': {
			{1, 1, 1, 0},
			{0, 0, 0, 1},
			{0, 1, 1, 0},
			{1, 0, 0, 0},
			{1, 1, 1, 1},
		},
		'3': {
			{1, 1, 1, 0},
			{0, 0, 0, 1},
			{0, 1, 1, 0},
			{0, 0, 0, 1},
			{1, 1, 1, 0},
		},
		'4': {
			{1, 0, 0, 1},
			{1, 0, 0, 1},
			{1, 1, 1, 1},
			{0, 0, 0, 1},
			{0, 0, 0, 1},
		},
		'5': {
			{1, 1, 1, 1},
			{1, 0, 0, 0},
			{1, 1, 1, 0},
			{0, 0, 0, 1},
			{1, 1, 1, 0},
		},
		'6': {
			{0, 1, 1, 1},
			{1, 0, 0, 0},
			{1, 1, 1, 0},
			{1, 0, 0, 1},
			{0, 1, 1, 0},
		},
		'7': {
			{1, 1, 1, 1},
			{0, 0, 0, 1},
			{0, 0, 1, 0},
			{0, 1, 0, 0},
			{1, 0, 0, 0},
		},
		'8': {
			{0, 1, 1, 0},
			{1, 0, 0, 1},
			{0, 1, 1, 0},
			{1, 0, 0, 1},
			{0, 1, 1, 0},
		},
		'9': {
			{0, 1, 1, 0},
			{1, 0, 0, 1},
			{0, 1, 1, 1},
			{0, 0, 0, 1},
			{1, 1, 1, 0},
		},
	}
	
	pattern, exists := patterns[char]
	if !exists {
		return
	}
	
	scale := 2 // 放大倍数
	for row, line := range pattern {
		for col, pixel := range line {
			if pixel == 1 {
				for dy := 0; dy < scale; dy++ {
					for dx := 0; dx < scale; dx++ {
						px := x + col*scale + dx
						py := y + row*scale + dy
						if px >= 0 && px < img.Bounds().Max.X && py >= 0 && py < img.Bounds().Max.Y {
							img.Set(px, py, c)
						}
					}
				}
			}
		}
	}
}

// addLines 添加干扰线
func (g *Generator) addLines(img *image.RGBA) {
	bounds := img.Bounds()
	
	for i := 0; i < 5; i++ {
		x1, _ := rand.Int(rand.Reader, big.NewInt(int64(bounds.Max.X)))
		y1, _ := rand.Int(rand.Reader, big.NewInt(int64(bounds.Max.Y)))
		x2, _ := rand.Int(rand.Reader, big.NewInt(int64(bounds.Max.X)))
		y2, _ := rand.Int(rand.Reader, big.NewInt(int64(bounds.Max.Y)))
		
		r, _ := rand.Int(rand.Reader, big.NewInt(256))
		g, _ := rand.Int(rand.Reader, big.NewInt(256))
		b, _ := rand.Int(rand.Reader, big.NewInt(256))
		
		lineColor := color.RGBA{uint8(r.Int64()), uint8(g.Int64()), uint8(b.Int64()), 255}
		
		drawLine(img, int(x1.Int64()), int(y1.Int64()), int(x2.Int64()), int(y2.Int64()), lineColor)
	}
}

// drawLine 绘制直线
func drawLine(img *image.RGBA, x1, y1, x2, y2 int, c color.RGBA) {
	dx := math.Abs(float64(x2 - x1))
	dy := math.Abs(float64(y2 - y1))
	
	sx := 1
	if x1 > x2 {
		sx = -1
	}
	
	sy := 1
	if y1 > y2 {
		sy = -1
	}
	
	err := dx - dy
	
	for {
		img.Set(x1, y1, c)
		
		if x1 == x2 && y1 == y2 {
			break
		}
		
		e2 := 2 * err
		if e2 > -dy {
			err -= dy
			x1 += sx
		}
		if e2 < dx {
			err += dx
			y1 += sy
		}
	}
}
