package matrix

import (
	"fmt"
	"gonum.org/v1/gonum/mat"
)

func toCustomMatrix(A *mat.Dense) *Matrix {
	rows, cols := A.Dims()
	data := make([][]float64, rows)

	// 提取矩阵数据
	for i := 0; i < rows; i++ {
		data[i] = make([]float64, cols)
		for j := 0; j < cols; j++ {
			data[i][j] = A.At(i, j)
		}
	}

	return &Matrix{
		Data: data,
		Rows: rows,
		Cols: cols,
	}
}

func toDenseMatrix(m *Matrix) *mat.Dense {
	// 创建一个包含 m.Rows * m.Cols 元素的切片
	data := make([]float64, m.Rows*m.Cols)

	// 按行填充数据
	for i := 0; i < m.Rows; i++ {
		for j := 0; j < m.Cols; j++ {
			data[i*m.Cols+j] = m.Data[i][j]
		}
	}

	// 创建并返回 mat.Dense 矩阵
	return mat.NewDense(m.Rows, m.Cols, data)
}

// SVD Jacobi 方法求特征值和特征向量
func SVD(a *Matrix) (*Matrix, *Matrix, *Matrix) {

	A := toDenseMatrix(a)

	// 计算 SVD 分解
	var svd mat.SVD
	ok := svd.Factorize(A, mat.SVDThin) // 计算 Thin SVD
	if !ok {
		fmt.Println("SVD 分解失败")
		return nil, nil, nil
	}

	// 获取 U、Σ、V^T
	var U, V mat.Dense
	svd.UTo(&U)
	svd.VTo(&V)

	// 获取奇异值 Σ
	values := svd.Values(nil)
	S := mat.NewDense(len(values), len(values), nil)
	for i, v := range values {
		S.Set(i, i, v) // 对角化 Σ 矩阵
	}
	return toCustomMatrix(&U), toCustomMatrix(S), toCustomMatrix(&V)
}

func Reconstruct(u, s, v *Matrix) *Matrix {
	U := toDenseMatrix(u)
	S := toDenseMatrix(s)
	V := toDenseMatrix(v)
	var US mat.Dense
	US.Mul(U, S) // 计算 U * Σ
	var AReconstructed mat.Dense
	AReconstructed.Mul(&US, V.T()) // 计算 (U * Σ) * V^T
	return toCustomMatrix(&AReconstructed)
}

//func SVD(a *Matrix) (*Matrix, *Matrix, *Matrix) {
//	epsilon := 1e-8
//	maxIter := 100
//
//	// 初始化 U, Sigma, V
//	u := NewMatrix(a.Rows, a.Cols)
//	sigma := NewMatrix(a.Rows, a.Cols)
//	v := NewMatrix(a.Cols, a.Cols)
//
//	// 初始化 V 为单位矩阵
//	for i := 0; i < v.Rows; i++ {
//		v.Data[i][i] = 1
//	}
//
//	// 拷贝 A 到 U
//	for i := 0; i < a.Rows; i++ {
//		for j := 0; j < a.Cols; j++ {
//			u.Data[i][j] = a.Data[i][j]
//		}
//	}
//
//	// 迭代
//	for iter := 0; iter < maxIter; iter++ {
//		// 遍历矩阵元素
//		maxOffDiagonal := 0.0
//		p, q := 0, 0
//		for i := 0; i < u.Rows; i++ {
//			for j := i + 1; j < u.Cols; j++ {
//				if math.Abs(u.Data[i][j]) > maxOffDiagonal {
//					maxOffDiagonal = math.Abs(u.Data[i][j])
//					p, q = i, j
//				}
//			}
//		}
//
//		// 检查收敛性
//		if maxOffDiagonal < epsilon {
//			break
//		}
//
//		// 计算旋转角度
//		theta := 0.5 * math.Atan2(2*u.Data[p][q], u.Data[p][p]-u.Data[q][q])
//
//		// 构造 Givens 旋转矩阵
//		c, s := math.Cos(theta), math.Sin(theta)
//
//		// 更新矩阵 U
//		for i := 0; i < u.Rows; i++ {
//			up := c*u.Data[i][p] - s*u.Data[i][q]
//			uq := s*u.Data[i][p] + c*u.Data[i][q]
//			u.Data[i][p] = up
//			u.Data[i][q] = uq
//		}
//
//		// 更新矩阵 V
//		for i := 0; i < v.Rows; i++ {
//			vp := c*v.Data[i][p] - s*v.Data[i][q]
//			vq := s*v.Data[i][p] + c*v.Data[i][q]
//			v.Data[i][p] = vp
//			v.Data[i][q] = vq
//		}
//	}
//
//	// 提取奇异值到 Sigma
//	for i := 0; i < sigma.Rows && i < sigma.Cols; i++ {
//		sigma.Data[i][i] = u.Data[i][i]
//	}
//
//	return u, sigma, v
//}

// Reconstruct 还原矩阵 A = U * Sigma * V^T
//func Reconstruct(U, Sigma, V *Matrix) *Matrix {
//	// 确保矩阵维度匹配
//	if U.Cols != Sigma.Rows || Sigma.Cols != V.Rows {
//		panic("矩阵维度不匹配")
//	}
//
//	// 初始化还原后的矩阵
//	result := NewMatrix(U.Rows, V.Cols)
//
//	// 进行矩阵乘法：A = U * Sigma
//	// 先计算 U * Sigma
//	USigma := NewMatrix(U.Rows, Sigma.Cols)
//	for i := 0; i < U.Rows; i++ {
//		for j := 0; j < Sigma.Cols; j++ {
//			USigma.Data[i][j] = 0
//			for k := 0; k < U.Cols; k++ {
//				USigma.Data[i][j] += U.Data[i][k] * Sigma.Data[k][j]
//			}
//		}
//	}
//
//	// 然后计算 (U * Sigma) * V^T
//	for i := 0; i < U.Rows; i++ {
//		for j := 0; j < V.Cols; j++ {
//			result.Data[i][j] = 0
//			for k := 0; k < Sigma.Cols; k++ {
//				result.Data[i][j] += USigma.Data[i][k] * V.Data[k][j]
//			}
//		}
//	}
//
//	return result
//}

//func main() {
//	// 示例矩阵
//	a := NewMatrix(3, 3)
//	a.data = [][]float64{
//		{1, 2, 3},
//		{4, 5, 6},
//		{7, 8, 9},
//	}
//
//	// 初始化 U, Σ, V 矩阵
//	u := NewMatrix(3, 3)
//	sigma := NewMatrix(3, 3)
//	v := NewMatrix(3, 3)
//
//	// 使用 Jacobi 方法进行 SVD
//	jacobiSVD(a, u, sigma, v)
//
//	// 打印结果
//	fmt.Println("矩阵 U:")
//	for _, row := range u.data {
//		fmt.Println(row)
//	}
//
//	fmt.Println("矩阵 Σ:")
//	for _, row := range sigma.data {
//		fmt.Println(row)
//	}
//
//	fmt.Println("矩阵 V:")
//	for _, row := range v.data {
//		fmt.Println(row)
//	}
//}
