package matrix

import (
	"math"
)

func dct(x []float64) []float64 {
	ret := make([]float64, len(x))
	for k := range x {
		for n := range x {
			q := math.Sqrt(1 / float64(len(x)))
			if k != 0 {
				q = math.Sqrt(2 / float64(len(x)))
			}
			ret[k] += q * x[n] * math.Cos(math.Pi*(float64(n)+0.5)*float64(k)/float64(len(x)))
		}
	}
	return ret
}

func idct(xk []float64) []float64 {
	ret := make([]float64, len(xk))
	for n := range xk {
		for k := range xk {
			q := math.Sqrt(1 / float64(len(xk)))
			if k != 0 {
				q = math.Sqrt(2 / float64(len(xk)))
			}
			ret[n] += q * xk[k] * math.Cos(math.Pi*(float64(n)+0.5)*float64(k)/float64(len(xk)))
		}
	}
	return ret
}

func Dct2(matrix Matrix) (*Matrix, error) {
	// 创建原始矩阵的副本
	x := make([][]float64, len(matrix.Data))
	for i := range matrix.Data {
		x[i] = append([]float64(nil), matrix.Data[i]...) // 创建每一行的副本
	}
	for i, v := range x {
		x[i] = dct(v)
	}
	x = switchRowAndColumns(x)
	for i, v := range x {
		x[i] = dct(v)
	}
	x = switchRowAndColumns(x)
	data, err := NewMatrixWithData(x)
	if err != nil {
		return nil, err
	}
	return data, nil
}

func IDct2(matrix *Matrix) (*Matrix, error) {
	x := matrix.Data
	x = switchRowAndColumns(x)
	for i, v := range x {
		x[i] = idct(v)
	}
	x = switchRowAndColumns(x)
	for i, v := range x {
		x[i] = idct(v)
	}
	data, err := NewMatrixWithData(x)
	if err != nil {
		return nil, err
	}
	return data, nil
}
