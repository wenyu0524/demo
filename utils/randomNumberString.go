package utils

import (
	"math/rand"
	"time"
)

// 生成随机数字字符串
func RandomNumberString(length int) string {
	r := rand.New(rand.NewSource(time.Now().UnixNano()))

	result := make([]byte, length)
	for i := range result {
		result[i] = byte('0' + r.Intn(10))
	}

	return string(result)
}
