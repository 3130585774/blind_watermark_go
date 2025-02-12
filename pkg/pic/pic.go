package pic

import (
	"errors"
	"fmt"
	"image"
	"image/color"
	_ "image/jpeg"
	"image/png"
	_ "image/png"
	"math"
	"os"
)

// ImageRGB 定义 ImageRGB 结构体
type ImageRGB struct {
	Width, Height int
	Pixels        [][]color.RGBA
}

// ImageYUV 定义 ImageYUV 结构体
type ImageYUV struct {
	Width, Height int
	Pixels        [][]YUV
}

// YUV 像素结构体
type YUV struct {
	Y, U, V uint8
}

// LoadFromFile 从文件读取 RGB 图片
func (rgb *ImageRGB) LoadFromFile(filePath string) error {
	file, err := os.Open(filePath)
	if err != nil {
		return err
	}
	defer func(file *os.File) {
		err := file.Close()
		if err != nil {
			panic(err)
		}
	}(file)

	// 读取图像
	srcImage, _, err := image.Decode(file)
	if err != nil {
		return err
	}

	// 初始化 ImageRGB 的宽、高和像素数据
	err = rgb.LoadFromImage(srcImage)
	if err != nil {
		return err
	}
	return nil
}

func (rgb *ImageRGB) LoadFromImage(image image.Image) error {
	bounds := image.Bounds()
	rgb.Width, rgb.Height = bounds.Dx(), bounds.Dy()

	// Initialize a 2D slice for Pixels
	rgb.Pixels = make([][]color.RGBA, rgb.Height)
	for i := range rgb.Pixels {
		rgb.Pixels[i] = make([]color.RGBA, rgb.Width)
	}

	for y := bounds.Min.Y; y < bounds.Max.Y; y++ {
		for x := bounds.Min.X; x < bounds.Max.X; x++ {
			rgba := color.RGBAModel.Convert(image.At(x, y)).(color.RGBA)
			err := rgb.Set(x, y, rgba)
			if err != nil {
				return err
			}
		}
	}
	return nil
}

func (rgb *ImageRGB) ToStandardImage() *image.RGBA {
	// 创建一个新的 RGBA 图像
	img := image.NewRGBA(image.Rect(0, 0, rgb.Width, rgb.Height))

	// 填充像素数据
	for y := 0; y < rgb.Height; y++ {
		for x := 0; x < rgb.Width; x++ {
			pixel := rgb.Pixels[y][x]
			img.SetRGBA(x, y, pixel)
		}
	}

	return img
}

func (rgb *ImageRGB) ToImageFile(fileName string) error {
	file, err := os.Create(fileName)
	if err != nil {
		panic(err)
	}
	defer func(file *os.File) {
		err := file.Close()
		if err != nil {
			fmt.Println(err.Error())
		}
	}(file)
	standardImage := rgb.ToStandardImage()
	err = png.Encode(file, standardImage)
	if err != nil {
		return err
	}
	return nil
}

// Set 设置单个像素（ImageRGB）
func (rgb *ImageRGB) Set(x, y int, c color.RGBA) error {
	if x < 0 || x >= rgb.Width || y < 0 || y >= rgb.Height {
		return errors.New("coordinate out of bounds")
	}
	rgb.Pixels[y][x] = c
	return nil
}

// Get 获取单个像素（ImageRGB）
func (rgb *ImageRGB) Get(x, y int) (color.RGBA, error) {
	if x < 0 || x >= rgb.Width || y < 0 || y >= rgb.Height {
		return color.RGBA{}, errors.New("coordinate out of bounds")
	}
	return rgb.Pixels[y][x], nil
}

// ToYUV 将 RGB 转换为 YUV
func (rgb *ImageRGB) ToYUV() *ImageYUV {
	yuv := &ImageYUV{
		Width:  rgb.Width,
		Height: rgb.Height,
	}

	// Initialize a 2D slice for YUV Pixels
	yuv.Pixels = make([][]YUV, rgb.Height)
	for i := range yuv.Pixels {
		yuv.Pixels[i] = make([]YUV, rgb.Width)
	}

	for h := 0; h < rgb.Height; h++ {
		for w := 0; w < rgb.Width; w++ {
			rgba := rgb.Pixels[h][w]
			r, g, b := rgba.R, rgba.G, rgba.B
			y, u, v := rgbToYUV(r, g, b)
			yuv.Pixels[h][w] = YUV{Y: y, U: u, V: v}
		}
	}
	return yuv
}

// Set 设置单个像素（ImageYUV）
func (yuv *ImageYUV) Set(x, y int, c YUV) error {
	if x < 0 || x >= yuv.Width || y < 0 || y >= yuv.Height {
		return errors.New("coordinate out of bounds")
	}
	yuv.Pixels[y][x] = c
	return nil
}

// Get 获取单个像素（ImageYUV）
func (yuv *ImageYUV) Get(x, y int) (YUV, error) {
	if x < 0 || x >= yuv.Width || y < 0 || y >= yuv.Height {
		return YUV{}, errors.New("coordinate out of bounds")
	}
	return yuv.Pixels[y][x], nil
}

// ToRGB 将 YUV 转换为 RGB
func (yuv *ImageYUV) ToRGB() *ImageRGB {
	rgb := &ImageRGB{
		Width:  yuv.Width,
		Height: yuv.Height,
	}

	// Initialize a 2D slice for RGB Pixels
	rgb.Pixels = make([][]color.RGBA, yuv.Height)
	for i := range rgb.Pixels {
		rgb.Pixels[i] = make([]color.RGBA, yuv.Width)
	}

	for y := 0; y < yuv.Height; y++ {
		for x := 0; x < yuv.Width; x++ {
			yuvPixel := yuv.Pixels[y][x]
			r, g, b := yuvToRGB(yuvPixel.Y, yuvPixel.U, yuvPixel.V)
			rgb.Pixels[y][x] = color.RGBA{R: r, G: g, B: b, A: 255}
		}
	}
	return rgb
}

// RGB 转 YUV 的转换公式
func rgbToYUV(r, g, b uint8) (y, u, v uint8) {
	y = uint8(0.299*float64(r) + 0.587*float64(g) + 0.114*float64(b))
	u = uint8(-0.14713*float64(r) - 0.28886*float64(g) + 0.436*float64(b) + 128)
	v = uint8(0.615*float64(r) - 0.51499*float64(g) - 0.10001*float64(b) + 128)
	return
}

// YUV 转 RGB 的转换公式
func yuvToRGB(y, u, v uint8) (r, g, b uint8) {
	y1 := float64(y)
	u1 := float64(u) - 128
	v1 := float64(v) - 128
	r = uint8(math.Max(0, math.Min(255, y1+1.13983*v1)))
	g = uint8(math.Max(0, math.Min(255, y1-0.39465*u1-0.58060*v1)))
	b = uint8(math.Max(0, math.Min(255, y1+2.03211*u1)))
	return
}
func ReadYUVWithData(yuvData [][][]float64) *ImageYUV {
	height := len(yuvData)
	if height == 0 {
		return nil
	}
	width := len(yuvData[0])

	// 初始化 ImageYUV 结构体
	imageYUV := &ImageYUV{
		Width:  width,
		Height: height,
		Pixels: make([][]YUV, height),
	}

	// 填充 Pixels 数组
	for y := 0; y < height; y++ {
		imageYUV.Pixels[y] = make([]YUV, width)
		for x := 0; x < width; x++ {
			imageYUV.Pixels[y][x] = YUV{
				Y: uint8(math.Max(0, math.Min(255, yuvData[y][x][0]))), // Y 分量
				U: uint8(math.Max(0, math.Min(255, yuvData[y][x][1]))), // U 分量
				V: uint8(math.Max(0, math.Min(255, yuvData[y][x][2]))), // V 分量
			}
		}
	}
	return imageYUV
}
