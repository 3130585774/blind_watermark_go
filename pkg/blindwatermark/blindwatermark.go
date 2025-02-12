package blind_watermark_go

import (
	"bytes"
	"encoding/base64"
	"fmt"
	"github.com/3130585774/blind_watermark_go/pkg/pic"
	"image"
	"image/png"
	"os"
	"strconv"
	"strings"
)

type WaterMark struct {
	BwmCore    *WaterMarkCore
	passwordWm int64
	WmBit      []bool
	wmSize     int
}

func NewWaterMark(blockShape [2]int) *WaterMark {

	return &WaterMark{
		BwmCore: NewWaterMarkCore(blockShape),
	}
}

func (wm *WaterMark) ReadImgFromBase64(encoded string) error {
	// 解码 base64 字符串
	data, err := base64.StdEncoding.DecodeString(encoded)
	if err != nil {
		return err
	}

	// 创建一个 io.Reader 来读取解码后的数据
	srcImage, _, err := image.Decode(bytes.NewReader(data))
	if err != nil {
		return err
	}

	// 处理解码后的图像数据
	wm.BwmCore.ReadImgArr(srcImage)
	return nil
}

func (wm *WaterMark) ReadImg(filename string) error {
	file, err := os.Open(filename)
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
	wm.BwmCore.ReadImgArr(srcImage)
	return nil
}

func (wm *WaterMark) ReadWm(wmContent string) error {
	return wm.BwmCore.ReadWm(wmContent)

}

func (wm *WaterMark) Embed(filename string) (*pic.ImageRGB, error) {
	embedImg := wm.BwmCore.Embed()
	if filename != "" {
		err := embedImg.ToImageFile(filename)
		if err != nil {
			return nil, err
		}
	}
	return embedImg, nil
}
func (wm *WaterMark) EmbedToBase64() (string, error) {
	embedImg := wm.BwmCore.Embed()
	var buf bytes.Buffer

	err := png.Encode(&buf, embedImg.ToStandardImage())
	if err != nil {
		return "", err
	}

	// 将字节流转换为 base64 字符串
	base64String := base64.StdEncoding.EncodeToString(buf.Bytes())

	return base64String, nil
}

func (wm *WaterMark) ExtractFromBase64(base64Str string) (string, error) {
	data, err := base64.StdEncoding.DecodeString(base64Str)
	if err != nil {
		return "", err
	}

	// 创建一个 io.Reader 来读取解码后的数据
	srcImage, _, err := image.Decode(bytes.NewReader(data))
	if err != nil {
		return "", err
	}
	r := wm.BwmCore.extractWithKMeans(srcImage, 256)
	fmt.Println(r)
	fmt.Println(len(r))
	decode, err := wm.BwmCore.bchDecode(r)
	if err != nil {
		return "", err
	}

	binaryString := boolArrayToBinaryString(decode)
	toBytes, err := binaryStringToBytes(binaryString)
	if err != nil {
		return "", err
	}
	return string(toBytes), nil
}

func (wm *WaterMark) Extract(filePath string) (string, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return "", err
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
		return "", err
	}
	r := wm.BwmCore.extractWithKMeans(srcImage, 256)
	fmt.Println(r)
	fmt.Println(len(r))
	decode, err := wm.BwmCore.bchDecode(r)
	if err != nil {
		return "", err
	}

	binaryString := boolArrayToBinaryString(decode)
	toBytes, err := binaryStringToBytes(binaryString)
	if err != nil {
		return "", err
	}
	return string(toBytes), nil
}

// 将布尔值切片转换为二进制字符串
func boolArrayToBinaryString(boolArray []bool) string {
	var binaryString strings.Builder
	for _, b := range boolArray {
		if b {
			binaryString.WriteString("1")
		} else {
			binaryString.WriteString("0")
		}
	}
	return binaryString.String()
}

// 将二进制字符串转换为字节数组
func binaryStringToBytes(binaryString string) ([]byte, error) {
	var byteArray []byte
	for i := 0; i < len(binaryString); i += 8 {
		// 获取每8位的二进制子串
		if i+8 <= len(binaryString) {
			bits := binaryString[i : i+8]
			byteValue, err := strconv.ParseUint(bits, 2, 8)
			if err != nil {
				return nil, err
			}
			byteArray = append(byteArray, byte(byteValue))
		}
	}
	return byteArray, nil
}
