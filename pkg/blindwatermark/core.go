package blind_watermark_go

import (
	"fmt"
	"github.com/3130585774/blind_watermark_go/pkg/matrix"
	"github.com/3130585774/blind_watermark_go/pkg/pic"
	"image"
	"image/color"
	"math"
	"math/rand"
	"strings"
	"sync"

	"github.com/zedseven/bch"
)

type WaterMarkCore struct {
	blockShape [2]int
	//passwordImg  int64
	d1, d2       float64
	img          pic.ImageRGB
	imgShape     [2]int
	imgYUV       pic.ImageYUV
	ca           [][][]float64
	hvd          [][][][]float64
	caShape      [2]int
	caBlock      [][][][][]float64
	caBlockShape [4]int
	caPart       [][][]float64
	wmBit        []bool
	wmSize       int
	blockNum     int
	partShape    [2]int
	blockIndex   [][2]int
	idxShuffle   [][]int
	bchConfig    *bch.EncodingConfig
}

func NewWaterMarkCore(blockshape [2]int) *WaterMarkCore {
	config, err := bch.CreateConfig(256, 10)
	if err != nil {
		panic(err)
	}
	return &WaterMarkCore{
		bchConfig:  config,
		blockShape: blockshape,
		d1:         30,
		d2:         12,
	}

}
func (wm *WaterMarkCore) InitBlockIndex() {

	wm.blockNum = wm.caBlockShape[0] * wm.caBlockShape[1]
	if wm.wmSize >= wm.blockNum {
		panic("水印大小超出容量")
	}
	wm.partShape = [2]int{wm.caBlockShape[0] * wm.blockShape[0], wm.caBlockShape[1] * wm.blockShape[1]}
	wm.blockIndex = make([][2]int, wm.blockNum)
	index := 0
	for i := 0; i < wm.caBlockShape[0]; i++ {
		for j := 0; j < wm.caBlockShape[1]; j++ {
			wm.blockIndex[index] = [2]int{i, j}
			index++
		}
	}
}

// 将 []bool 转换为 []int
func boolToInt(b []bool) []byte {
	result := make([]byte, len(b))
	for i, v := range b {
		if v {
			result[i] = 1
		} else {
			result[i] = 0
		}
	}
	return result
}

// 将 []int 转换为 []bool
func intToBool(b []byte) []bool {
	result := make([]bool, len(b))
	for i, v := range b {
		result[i] = v == 1
	}
	return result
}

func (wm *WaterMarkCore) bchEncode(infoBits []bool) ([]bool, error) {
	// 将输入的布尔信息位转换为整数数组
	input := boolToInt(infoBits)
	//config, err := bch.CreateConfig(len(input), len(input)/2)

	encoded, err := bch.Encode(wm.bchConfig, &input)
	if err != nil {
		return nil, err
	}

	return intToBool(encoded), nil
}

func (wm *WaterMarkCore) bchDecode(encodedBits []bool) ([]bool, error) {
	// 将输入的布尔数组转换为整数数组
	// 将输入的布尔信息位转换为整数数组
	input := boolToInt(encodedBits)

	encoded, _, err := bch.Decode(wm.bchConfig, &input)
	if err != nil {
		return nil, err
	}

	return intToBool(encoded), nil
}

// ReadWm 读取水印
func (wm *WaterMarkCore) ReadWm(wmStr string) error {

	// 转字符串为二进制比特数组
	byteData := []byte(wmStr)
	toBinaryString := bytesToBinaryString(byteData)

	//wm.wmBit = binaryStringToBoolArray(toBinaryString)[1:]

	rawBit := binaryStringToBoolArray(toBinaryString)

	encode, err := wm.bchEncode(rawBit)

	if err != nil {
		return err
	}

	wm.wmBit = encode
	fmt.Println(wm.wmBit)
	fmt.Println(len(wm.wmBit))
	wm.wmSize = len(wm.wmBit)

	return nil
}

// 将字节数组转换为二进制字符串
func bytesToBinaryString(data []byte) string {
	var binaryString strings.Builder
	for _, b := range data {
		// fmt.Sprintf("%08b", b) 将字节转为8位二进制字符串
		binaryString.WriteString(fmt.Sprintf("%08b", b))
	}
	return binaryString.String()
}

// 将二进制字符串转换为布尔值切片
func binaryStringToBoolArray(binaryString string) []bool {
	boolArray := make([]bool, len(binaryString))
	for i, c := range binaryString {
		if c == '1' {
			boolArray[i] = true
		} else {
			boolArray[i] = false
		}
	}
	return boolArray
}

// shuffleBitArray 随机打乱比特数组
func (wm *WaterMarkCore) shuffleBitArray(seed int64) {
	r := rand.New(rand.NewSource(seed))
	r.Shuffle(len(wm.wmBit), func(i, j int) {
		wm.wmBit[i], wm.wmBit[j] = wm.wmBit[j], wm.wmBit[i]
	})
}

// imageToGray 将图片转为灰度图
func imageToGray(img image.Image) *image.Gray {
	bounds := img.Bounds()
	grayImg := image.NewGray(bounds)
	for y := bounds.Min.Y; y < bounds.Max.Y; y++ {
		for x := bounds.Min.X; x < bounds.Max.X; x++ {
			c := img.At(x, y)
			r, g, b, _ := c.RGBA()
			gray := uint8((r*299 + g*587 + b*114 + 500) / 1000 >> 8)
			grayImg.SetGray(x, y, color.Gray{Y: gray})
		}
	}
	return grayImg
}

// grayToBitArray 将灰度图片转换为比特数组
func grayToBitArray(img *image.Gray) []bool {
	bounds := img.Bounds()
	bitArray := make([]bool, bounds.Dx()*bounds.Dy())
	index := 0
	for y := bounds.Min.Y; y < bounds.Max.Y; y++ {
		for x := bounds.Min.X; x < bounds.Max.X; x++ {
			if img.GrayAt(x, y).Y > 128 {
				bitArray[index] = true
			} else {
				bitArray[index] = false
			}
			index++
		}
	}
	return bitArray
}

func (wm *WaterMarkCore) ReadImgArr(img image.Image) {

	//读入图片->YUV化->加白边使像素变偶数->四维分块
	err := wm.img.LoadFromImage(img)
	if err != nil {
		panic(err)
	}

	wm.imgShape = [2]int{wm.img.Height, wm.img.Width}

	//YUV化
	wm.ConvertToYUV()

	//加白边使像素变偶数
	wm.AddWhiteBorderToMakeEvenDimensions()

	//四维分块
	for i := 0; i < 2; i++ {
		wm.caShape[i] = (wm.imgShape[i] + 1) / 2
	}

	// 计算 ca_block_shape
	wm.caBlockShape[0] = wm.caShape[0] / wm.blockShape[0]
	wm.caBlockShape[1] = wm.caShape[1] / wm.blockShape[1]
	wm.caBlockShape[2] = wm.blockShape[0]
	wm.caBlockShape[3] = wm.blockShape[1]

	// 计算 strides
	var strides [4]int
	//strides[0] = 4 * wm.caShape[1] * wm.blockShape[0]
	//strides[1] = 4 * wm.blockShape[1]
	//strides[2] = 4 * wm.caShape[1]
	//strides[3] = 4
	strides[0] = wm.caShape[1] * wm.blockShape[0]
	strides[1] = wm.blockShape[1]
	strides[2] = wm.caShape[1]
	strides[3] = 1

	imgMatrix := wm.extractChannel(wm.imgYUV.Pixels)
	for channel := 0; channel < 3; channel++ {
		ma, _ := matrix.NewMatrixWithData(imgMatrix[channel])
		ca, ch, cv, cd := matrix.Dwt2(ma)
		if wm.ca == nil {
			wm.ca = make([][][]float64, 3) // 初始化 ca 切片
		}
		wm.ca[channel] = ca.Data

		if wm.hvd == nil {
			wm.hvd = make([][][][]float64, 3) // 初始化 hvd 切片
		}
		wm.hvd[channel] = make([][][]float64, 3) // 每个通道有 3 个部分

		wm.hvd[channel][0] = ch.Data
		wm.hvd[channel][1] = cv.Data
		wm.hvd[channel][2] = cd.Data

		if wm.caBlock == nil {
			wm.caBlock = make([][][][][]float64, 3)
		}

		wm.caBlock[channel] = asStrided4D(wm.ca[channel], wm.caBlockShape, strides)
	}

}

func (wm *WaterMarkCore) extractChannel(imgYuv [][]pic.YUV) (result [][][]float64) {
	if len(imgYuv) == 0 || len(imgYuv[0]) == 0 {
		return nil
	}
	rows := len(imgYuv)
	cols := len(imgYuv[0])
	result = make([][][]float64, 3)
	for i := 0; i < 3; i++ {
		result[i] = make([][]float64, rows)
		for j := 0; j < rows; j++ {
			result[i][j] = make([]float64, cols)
		}
	}
	for channel := 0; channel < 3; channel++ {
		for h := 0; h < rows; h++ {
			for w := 0; w < cols; w++ {
				var value float64
				if channel == 0 {
					value = float64(imgYuv[h][w].Y)
				}
				if channel == 1 {
					value = float64(imgYuv[h][w].U)
				}
				if channel == 2 {
					value = float64(imgYuv[h][w].V)
				}
				result[channel][h][w] = value
			}
		}
	}
	return result
}

func (wm *WaterMarkCore) combineChannel(result [][][]float64) [][]pic.YUV {
	if len(result) == 0 || len(result[0]) == 0 || len(result[0][0]) == 0 {
		return nil
	}

	// 获取维度信息
	channels := len(result)
	rows := len(result[0])
	cols := len(result[0][0])

	if channels != 3 {
		panic("Input must have exactly 3 channels (Y, U, V)")
	}

	// 初始化二维 YUV 图像数组
	imgYuv := make([][]pic.YUV, rows)
	for i := 0; i < rows; i++ {
		imgYuv[i] = make([]pic.YUV, cols)
	}

	// 遍历填充回 YUV 格式
	for h := 0; h < rows; h++ {
		for w := 0; w < cols; w++ {
			imgYuv[h][w] = pic.YUV{
				Y: uint8(result[0][h][w]),
				U: uint8(result[1][h][w]),
				V: uint8(result[2][h][w]),
			}
		}
	}

	return imgYuv
}

func (wm *WaterMarkCore) ConvertToYUV() {
	wm.imgYUV = *wm.img.ToYUV()
}

// AddWhiteBorderToMakeEvenDimensions 添加白色边框，确保宽度和高度都是偶数
func (wm *WaterMarkCore) AddWhiteBorderToMakeEvenDimensions() {
	// 如果宽度是奇数，则增加一个像素宽度的边框
	originalWidth := wm.imgYUV.Width
	originalHeight := wm.imgYUV.Height

	if wm.imgYUV.Width%2 != 0 {
		wm.imgYUV.Width++
	}

	// 如果高度是奇数，则增加一个像素高度的边框
	if wm.imgYUV.Height%2 != 0 {
		wm.imgYUV.Height++
	}

	// 计算原图像的宽度和高度

	newPixels := make([][]pic.YUV, wm.imgYUV.Height) // 第一维是行数（Height）
	for i := range newPixels {
		newPixels[i] = make([]pic.YUV, wm.imgYUV.Width) // 每行分配列数（Width）
	}

	// 填充原始图像部分的 YUV 数据
	for h := 0; h < originalHeight; h++ {
		for w := 0; w < originalWidth; w++ {
			newPixels[h][w] = wm.imgYUV.Pixels[h][w]
		}
	}

	// 填充白色边框
	whitePixel := pic.YUV{Y: 255, U: 128, V: 128}

	for h := 0; h < wm.imgYUV.Height; h++ {
		for w := 0; w < wm.imgYUV.Width; w++ {
			if w < originalWidth && h < originalHeight {
				continue // 不修改原始图像区域
			}
			newPixels[h][w] = whitePixel
		}
	}
	wm.imgYUV.Pixels = newPixels
}

func asStrided4D(arr [][]float64, shape, strides [4]int) [][][][]float64 {
	//FIXME 逻辑有问题
	d0, d1, d2, d3 := shape[0], shape[1], shape[2], shape[3]
	numElements := d0 * d1 * d2 * d3

	// 创建一个结果数组
	result := make([][][][]float64, d0)
	for i0 := 0; i0 < d0; i0++ {
		result[i0] = make([][][]float64, d1)
		for i1 := 0; i1 < d1; i1++ {
			result[i0][i1] = make([][]float64, d2)
			for i2 := 0; i2 < d2; i2++ {
				result[i0][i1][i2] = make([]float64, d3)
			}
		}
	}

	// 计算每个元素的偏移量并访问原数组
	for i := 0; i < numElements; i++ {
		// 计算当前元素的多维索引
		multiIndex := make([]int, 4)
		remainingIndex := i
		for j := 3; j >= 0; j-- {
			multiIndex[j] = remainingIndex % shape[j]
			remainingIndex /= shape[j]
		}

		// 计算偏移量
		offset := multiIndex[0]*strides[0] + multiIndex[1]*strides[1] + multiIndex[2]*strides[2] + multiIndex[3]*strides[3]

		// 确保偏移在原数组的有效范围内
		a := len(arr) * len(arr[0])
		if offset >= a || offset < 0 {
			return nil
		}

		// 计算在原二维数组中的行列位置
		row := offset / len(arr[0])
		col := offset % len(arr[0])

		// 将原数组中的值放到结果视图中
		result[multiIndex[0]][multiIndex[1]][multiIndex[2]][multiIndex[3]] = arr[row][col]
	}
	return result
}

func (wm *WaterMarkCore) Embed() *pic.ImageRGB {
	wm.InitBlockIndex()

	// DeepCopy embedCA
	embedCA := make([][][]float64, len(wm.ca))
	for i := range wm.ca {
		embedCA[i] = make([][]float64, len(wm.ca[i]))
		for j := range wm.ca[i] {
			embedCA[i][j] = make([]float64, len(wm.ca[i][j]))
			copy(embedCA[i][j], wm.ca[i][j])
		}
	}

	embedYUV := make([][][]float64, 3)

	//wm.idxShuffle = RandomStrategy(wm.passwordImg, wm.blockNum, wm.blockShape[0]*wm.blockShape[1])

	//var wg sync.WaitGroup
	tmp := make([][][]float64, wm.blockNum)

	for channel := 0; channel < 3; channel++ {
		var wg sync.WaitGroup
		wg.Add(wm.blockNum)
		for i := 0; i < wm.blockNum; i++ {
			go func(i int) {
				defer wg.Done()
				tmp[i] = wm.BlockAddWM(wm.caBlock[channel][wm.blockIndex[i][0]][wm.blockIndex[i][1]], i)
			}(i)
		}
		wg.Wait()

		//for i := 0; i < wm.blockNum; i++ {
		//	// 同步执行任务，按顺序操作
		//	tmp[i] = wm.BlockAddWM(wm.caBlock[channel][wm.blockIndex[i][0]][wm.blockIndex[i][1]], i)
		//}

		// Update blocks
		for i := 0; i < wm.blockNum; i++ {
			wm.caBlock[channel][wm.blockIndex[i][0]][wm.blockIndex[i][1]] = tmp[i]
		}

		// Merge 4D blocks back into 2D
		wm.caPart = make([][][]float64, 3)

		wm.caPart[channel] = ConcatenateBlocks(wm.caBlock[channel])

		// Update embedCA
		for x := 0; x < wm.partShape[0]; x++ {
			for y := 0; y < wm.partShape[1]; y++ {
				embedCA[channel][x][y] = wm.caPart[channel][x][y]
			}
		}

		// Perform inverse transform
		embedYUV[channel] = matrix.IDwt2(embedCA[channel], wm.hvd[channel][0], wm.hvd[channel][1], wm.hvd[channel][2])

	}

	// Combine YUV channels
	embedImgYUV := StackChannels(embedYUV)

	// Trim to original image shape
	embed := TrimImage(embedImgYUV, wm.imgShape)

	imageYuv := pic.ReadYUVWithData(embed)
	// Convert to BGR
	embedImg := imageYuv.ToRGB()

	return embedImg
}

func TrimImage(yuv [][][]float64, shape [2]int) [][][]float64 {
	// 创建一个新的二维切片来存储裁剪后的图像
	if (len(yuv[0]) == shape[0]) && len(yuv) == shape[1] {
		return yuv
	}
	trimmedImage := make([][][]float64, shape[0]) // 复制 yuv 的第一级切片数量

	// 遍历每个 channel
	for x := 0; x < shape[0]; x++ {
		trimmedImage[x] = make([][]float64, shape[1]) // 为每个 channel 创建适当的大小
		for y := 0; y < shape[1]; y++ {
			trimmedImage[x][y] = make([]float64, 3) // 列数为 shape[1]
			copy(trimmedImage[x][y], yuv[x][y])     // 将部分数据复制到新切片
		}
	}
	return trimmedImage
}

func StackChannels(yuv [][][]float64) [][][]float64 {
	// 确保 Y, U, V 通道的行列数相同
	rows := len(yuv[0])
	cols := len(yuv[0][0])

	// 创建一个三维切片，用来存储合并后的 YUV 图像
	embedImgYUV := make([][][]float64, rows)

	for i := 0; i < rows; i++ {
		embedImgYUV[i] = make([][]float64, cols)
		for j := 0; j < cols; j++ {
			// 合并 Y, U, V 通道为三维数组
			embedImgYUV[i][j] = []float64{
				yuv[0][i][j], // Y 通道
				yuv[1][i][j], // U 通道
				yuv[2][i][j], // V 通道
			}
		}
	}

	return embedImgYUV
}

func ConcatenateBlocks(channelBlocks [][][][]float64) [][]float64 {

	// 获取块的维度信息
	blockRows := len(channelBlocks)           // 块的行数
	blockCols := len(channelBlocks[0])        // 块的列数
	blockHeight := len(channelBlocks[0][0])   // 每块的高度
	blockWidth := len(channelBlocks[0][0][0]) // 每块的宽度

	// 初始化合并后的 2D 矩阵
	totalRows := blockRows * blockHeight
	totalCols := blockCols * blockWidth
	merged := make([][]float64, totalRows)
	for i := range merged {
		merged[i] = make([]float64, totalCols)
	}

	// 遍历块并填充到结果数组
	for blockRow := 0; blockRow < blockRows; blockRow++ {
		for blockCol := 0; blockCol < blockCols; blockCol++ {
			for i := 0; i < blockHeight; i++ {
				for j := 0; j < blockWidth; j++ {
					// 计算全局位置
					globalRow := blockRow*blockHeight + i
					globalCol := blockCol*blockWidth + j
					merged[globalRow][globalCol] = channelBlocks[blockRow][blockCol][i][j]
				}
			}
		}
	}

	return merged
}

func (wm *WaterMarkCore) BlockAddWM(block [][]float64, i int) [][]float64 {
	//if !wm.fastMode {
	return wm.blockAddWMFast(block, i)
	//} else {
	//	return wm.blockAddWMSlow(block, shuffler, i)
	//}
}

// blockAddWMSlow performs watermark embedding with more steps (slow mode).
func (wm *WaterMarkCore) blockAddWMSlow(block [][]float64, shuffler []int, i int) [][]float64 {
	wm1 := wm.wmBit[i%wm.wmSize]

	data, err := matrix.NewMatrixWithData(block)
	if err != nil {
		return nil
	}
	// Apply DCT
	dct2, err := matrix.Dct2(*data)
	if err != nil {
		return nil
	}

	// Shuffling DCT coefficients
	dct2Flatten := dct2.Flatten()
	shuffled := make([]float64, len(dct2Flatten))
	for j, idx := range shuffler {
		shuffled[j] = dct2Flatten[idx]
	}

	rows, cols := wm.blockShape[0], wm.blockShape[1]
	reshaped := make([][]float64, rows)
	for j := 0; j < rows; j++ {
		reshaped[j] = shuffled[j*cols : (j+1)*cols]
	}

	matrix1, err := matrix.NewMatrixWithData(reshaped)
	if err != nil {
		return nil
	}

	// Apply SVD
	u, s, v := matrix.SVD(matrix1)

	// Modify singular values with watermark
	var wm1Value float64
	if wm1 {
		wm1Value = 1.0
	} else {
		wm1Value = 0.0
	}

	// Modify the first singular value
	s.Data[0][0] = math.Floor(s.Data[0][0]/wm.d1+0.25+0.5*wm1Value) * wm.d1
	if wm.d2 > 0 {
		s.Data[1][1] = math.Floor(s.Data[1][1]/wm.d2+0.25+0.5*wm1Value) * wm.d2
	}

	// Reconstruct the matrix
	reconstructed := matrix.Reconstruct(u, s, v)

	blockDctFlatten := reconstructed.Flatten()

	blockCopy := make([]float64, len(blockDctFlatten))
	copy(blockCopy, blockDctFlatten)

	// Step 2: Rearrange the original array using the shuffler
	for i, idx := range shuffler {
		blockDctFlatten[i] = blockCopy[idx]
	}

	blockDct := make([][]float64, rows)
	for j := 0; j < rows; j++ {
		blockDct[j] = blockDctFlatten[j*cols : (j+1)*cols]
	}
	// Restore to original dimensions

	matrix2, err := matrix.NewMatrixWithData(blockDct)
	if err != nil {
		return nil
	}
	iDct2, err := matrix.IDct2(matrix2)
	if err != nil {
		return nil
	}
	return iDct2.Data
}

// blockAddWMFast performs watermark embedding in a faster way (fast mode).
func (wm *WaterMarkCore) blockAddWMFast(block [][]float64, i int) [][]float64 {
	wm1 := wm.wmBit[i%wm.wmSize]

	data, err := matrix.NewMatrixWithData(block)
	if err != nil {
		return nil
	}
	// Apply DCT
	dct2, err := matrix.Dct2(*data)
	if err != nil {
		return nil
	}

	u, s, v := matrix.SVD(dct2)
	// Apply SVD (simulated here, use real SVD in your implementation)

	// Modify singular values with watermark
	var wm1Value float64
	if wm1 {
		wm1Value = 1.0
	} else {
		wm1Value = 0.0
	}

	s.Data[0][0] = (s.Data[0][0]/wm.d1 + 0.25 + 0.5*wm1Value) * wm.d1

	reconstruct := matrix.Reconstruct(u, s, v)

	iDct2, err := matrix.IDct2(reconstruct)
	if err != nil {
		return nil
	}
	return iDct2.Data
}

// 计算一个切片的均值
func mean(data []float64) float64 {
	var sum float64
	for _, val := range data {
		sum += val
	}
	return sum / float64(len(data))
}

// 一维K-means聚类
func oneDimKMeans(inputs []float64) []bool {
	eTol := 1e-6
	center := []float64{inputs[0], inputs[len(inputs)-1]} // 1. 初始化中心点
	var threshold float64
	for i := 0; i < 10000; i++ {
		threshold = (center[0] + center[1]) / 2
		// 2. 检查所有点与这两个中心点之间的距离，并将每个点归类到最近的中心
		isClass01 := make([]bool, len(inputs))
		for j, value := range inputs {
			isClass01[j] = value > threshold
		}

		// 3. 重新计算中心点
		var class0, class1 []float64
		for j, value := range inputs {
			if isClass01[j] {
				class1 = append(class1, value)
			} else {
				class0 = append(class0, value)
			}
		}
		center[0] = mean(class0)
		center[1] = mean(class1)

		// 4. 判断停止条件
		if math.Abs((center[0]+center[1])/2-threshold) < eTol {
			threshold = (center[0] + center[1]) / 2
			break
		}
	}

	// 返回最终分类
	isClass01 := make([]bool, len(inputs))
	for j, value := range inputs {
		isClass01[j] = value > threshold
	}
	return isClass01
}

// 从图片中提取水印

func (wm *WaterMarkCore) blockGetWm(block [][]float64) (float64, error) {
	return wm.blockGetWmFast(block)
}

func (wm *WaterMarkCore) blockGetWmFast(block [][]float64) (float64, error) {

	data, err := matrix.NewMatrixWithData(block)
	if err != nil {
		return 0, err
	}
	// dct->svd->解水印
	dct2, err := matrix.Dct2(*data)
	if err != nil {
		return 0, err
	}
	_, s, _ := matrix.SVD(dct2)

	watermark := 0
	if math.Mod(s.Data[0][0], wm.d1) > wm.d1/2 {
		watermark = 1
	}
	return float64(watermark), nil
}

func (wm *WaterMarkCore) extractRaw(img image.Image) ([][]float64, error) {
	// 每个分块提取1 bit信息
	wm.ReadImgArr(img)
	wm.InitBlockIndex()

	wmBlockBit := make([][]float64, 3)
	for channel := 0; channel < 3; channel++ {
		wmBlockBit[channel] = make([]float64, wm.blockNum)
	}

	//wm.idxShuffle = randomStrategy1(wm.passwordImg, wm.blockNum, wm.blockShape[0]*wm.blockShape[1])
	//var err error
	for channel := 0; channel < 3; channel++ {
		var wg sync.WaitGroup
		wg.Add(wm.blockNum)
		for i := 0; i < wm.blockNum; i++ {
			go func(i int) {
				defer wg.Done()
				wmBlockBit[channel][i], _ = wm.blockGetWm(wm.caBlock[channel][wm.blockIndex[i][0]][wm.blockIndex[i][1]])
			}(i)
		}
		wg.Wait()

		//for i := 0; i < wm.blockNum; i++ {
		//	wmBlockBit[channel][i], err = wm.blockGetWm(wm.caBlock[channel][wm.blockIndex[i][0]][wm.blockIndex[i][1]])
		//	if err != nil {
		//		return nil, err
		//	}
		//}
	}
	return wmBlockBit, nil
}

//func randomStrategy1(img int64, num int, i int) [][]int {
//
//}

func (wm *WaterMarkCore) extractAvg(wmBlockBit [][]float64) []float64 {
	// 初始化 wmAvg 切片，长度为 wmSize
	wmAvg := make([]float64, wm.wmSize)

	// 遍历每个起始列索引 i（即 0 到 wmSize-1）
	for i := 0; i < wm.wmSize; i++ {
		var sum float64
		count := 0

		// 遍历所有行
		for row := 0; row < len(wmBlockBit); row++ {
			//for row := 0; row < 1; row++ {
			// 从索引 i 开始，以 wmSize 作为步长遍历列
			for col := i; col < len(wmBlockBit[row]); col += wm.wmSize {
				sum += wmBlockBit[row][col]
				count++
			}
		}

		// 计算平均值
		if count > 0 {
			wmAvg[i] = sum / float64(count)
		}
	}

	return wmAvg
}

func (wm *WaterMarkCore) extract(img image.Image, wmLen int) []float64 {
	wm.wmSize = wmLen

	// 提取每个分块埋入的bit：
	wmBlockBit, _ := wm.extractRaw(img)

	// 做平均：
	wmAvg := wm.extractAvg(wmBlockBit)

	return wmAvg
}

func (wm *WaterMarkCore) extractWithKMeans(img image.Image, wmLen int) []bool {
	wmAvg := wm.extract(img, wmLen)
	return oneDimKMeans(wmAvg)
}
