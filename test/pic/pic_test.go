package pic

import (
	pic "github.com/3130585774/blind_watermark_go/pkg/pic"
	"image"
	"image/color"
	_ "image/jpeg"
	_ "image/png"
	"os"
	"testing"
)

// TestLoadFromImage 测试 LoadFromImage 函数
func TestLoadFromImage(t *testing.T) {
	file, err := os.Open("./test.png")
	if err != nil {
		t.Fatalf("failed to open file: %v", err)
	}
	defer func(file *os.File) {
		err := file.Close()
		if err != nil {
			t.Fatalf("failed to close file: %v", err)
		}
	}(file)

	img, _, err := image.Decode(file)
	if err != nil {
		t.Fatalf("failed to decode image: %v", err)
	}

	rgb := &pic.ImageRGB{}
	err = rgb.LoadFromFile("./test.png")
	if err != nil {
		t.Fatalf("Failed to load image: %v", err)
	}

	if rgb.Width != img.Bounds().Dx() || rgb.Height != img.Bounds().Dy() {
		t.Fatalf("Image dimensions do not match: expected %dx%d, got %dx%d", img.Bounds().Dx(), img.Bounds().Dy(), rgb.Width, rgb.Height)
	}

	yuv := rgb.ToYUV()
	err = yuv.ToRGB().ToImageFile("./test2.png")
	if err != nil {
		t.Fatalf("Failed to convert image: %v", err)
	}

}

func TestMakeAWhitePicture(t *testing.T) {

	p := &pic.ImageYUV{
		Width:  100,
		Height: 100,
	}
	newPixels := make([][]pic.YUV, p.Height) // 第一维是行数（Height）
	for i := range newPixels {
		newPixels[i] = make([]pic.YUV, p.Width) // 每行分配列数（Width）
	}

	y, u, v := rgbToYUV(255, 255, 255)

	wyun := pic.YUV{Y: y, U: u, V: v}

	for w := 0; w < p.Width; w++ {
		for h := 0; h < p.Height; h++ {
			newPixels[h][w] = wyun
		}
	}
	p.Pixels = newPixels
	err := p.ToRGB().ToImageFile("./test3.png")
	if err != nil {
		t.Fatalf("Failed to convert image: %v", err)
	}
}
func rgbToYUV(r, g, b uint8) (y, u, v uint8) {
	y = uint8(0.299*float64(r) + 0.587*float64(g) + 0.114*float64(b))
	u = uint8(-0.14713*float64(r) - 0.28886*float64(g) + 0.436*float64(b) + 128)
	v = uint8(0.615*float64(r) - 0.51499*float64(g) - 0.10001*float64(b) + 128)
	return
}

// TestSetAndGetRGB 测试 Set 和 Get 方法
func TestSetAndGetRGB(t *testing.T) {
	rgb := &pic.ImageRGB{
		Width:  2,
		Height: 2,
		Pixels: make([][]color.RGBA, 2),
	}
	for i := range rgb.Pixels {
		rgb.Pixels[i] = make([]color.RGBA, 2)
	}

	err := rgb.Set(0, 0, color.RGBA{R: 255, G: 128, B: 64, A: 255})
	if err != nil {
		t.Fatalf("Failed to set pixel: %v", err)
	}

	c, err := rgb.Get(0, 0)
	if err != nil {
		t.Fatalf("Failed to get pixel: %v", err)
	}

	if c.R != 255 || c.G != 128 || c.B != 64 || c.A != 255 {
		t.Errorf("Incorrect pixel value: got %v, expected {255 128 64 255}", c)
	}
}
