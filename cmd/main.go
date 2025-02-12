package main

import (
	"fmt"
	bwm "github.com/3130585774/blind_watermark_go/pkg/blindwatermark"
)

func main() {
	// Example usage
	wm := bwm.NewWaterMark([2]int{2, 2})

	// Read image and watermark
	err := wm.ReadImg("./ori_img.jpeg")
	if err != nil {
		fmt.Println(err)
		return
	}

	// Read watermark
	err = wm.ReadWm("123456789012345678901234567890")
	if err != nil {
		fmt.Println(err)
		return
	}

	// Embed watermark into image and save
	_, err = wm.Embed("go.png")
	if err != nil {
		fmt.Println(err)
		return
	}

	//Extract watermark
	wm2 := bwm.NewWaterMark([2]int{2, 2})
	extractedWm, err := wm2.Extract("go.png")
	if err != nil {
		fmt.Println(err)
		return
	}
	fmt.Println(extractedWm)

}
