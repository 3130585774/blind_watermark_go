package matrix

import "fmt"

// Matrix 矩阵定义
type Matrix struct {
	Data       [][]float64
	Rows, Cols int
}

// NewMatrix 初始化矩阵
func NewMatrix(rows, cols int) *Matrix {
	data := make([][]float64, rows)
	for i := range data {
		data[i] = make([]float64, cols)
	}
	return &Matrix{Data: data, Rows: rows, Cols: cols}
}

// NewMatrixWithData 根据已有数据创建矩阵
func NewMatrixWithData(data [][]float64) (*Matrix, error) {
	// 检查数据是否有效（每一行的列数是否一致）
	if len(data) == 0 {
		return nil, fmt.Errorf("data cannot be empty")
	}
	cols := len(data[0])
	for i := range data {
		if len(data[i]) != cols {
			return nil, fmt.Errorf("all rows must have the same number of columns")
		}
	}
	return &Matrix{Data: data, Rows: len(data), Cols: cols}, nil
}

// Multiply 矩阵乘法
func (m *Matrix) Multiply(n *Matrix) *Matrix {
	if m.Cols != n.Rows {
		panic("矩阵维度不匹配")
	}
	result := NewMatrix(m.Rows, n.Cols)
	for i := 0; i < m.Rows; i++ {
		for j := 0; j < n.Cols; j++ {
			sum := 0.0
			for k := 0; k < m.Cols; k++ {
				sum += m.Data[i][k] * n.Data[k][j]
			}
			result.Data[i][j] = sum
		}
	}
	return result
}

// Transpose 转置矩阵
func (m *Matrix) Transpose() *Matrix {
	result := NewMatrix(m.Cols, m.Rows)
	for i := 0; i < m.Rows; i++ {
		for j := 0; j < m.Cols; j++ {
			result.Data[j][i] = m.Data[i][j]
		}
	}
	return result
}

func (m *Matrix) Flatten() []float64 {
	var result []float64

	// 遍历二维矩阵，将元素逐个添加到结果切片中
	for _, row := range m.Data {
		result = append(result, row...)
	}

	return result
}
