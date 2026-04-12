package quanyu

import (
	"crypto/md5"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"math/big"
	"strings"
	"time"
)

// GenerateNonceStr 生成随机字符串
func GenerateNonceStr(length int) string {
	const charset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	b := make([]byte, length)
	for i := range b {
		n, _ := rand.Int(rand.Reader, big.NewInt(int64(len(charset))))
		b[i] = charset[n.Int64()]
	}
	return string(b)
}

// GetCurrentTimestamp 获取当前时间戳 格式: yyyyMMddHHmmss
func GetCurrentTimestamp() string {
	return time.Now().Format("20060102150405")
}

// GenerateSign 生成签名
// 签名规则: sign = MD5("appid={appid}&nonce_str={nonce_str}&uid={uid}&key={key}").toUpperCase()
func GenerateSign(appid, nonceStr, uid, key string) string {
	signStr := fmt.Sprintf("appid=%s&nonce_str=%s&uid=%s&key=%s", appid, nonceStr, uid, key)

	hash := md5.New()
	hash.Write([]byte(signStr))
	hashBytes := hash.Sum(nil)

	return strings.ToUpper(hex.EncodeToString(hashBytes))
}
