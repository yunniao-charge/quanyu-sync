package quanyu

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGenerateSign(t *testing.T) {
	appid := "test_appid"
	nonceStr := "abc123"
	uid := "TEST_UID"
	key := "test_key"

	sign := GenerateSign(appid, nonceStr, uid, key)

	assert.Len(t, sign, 32, "签名长度应为32")
	assert.Equal(t, strings.ToUpper(sign), sign, "签名应为大写")
}

func TestGenerateSignDeterministic(t *testing.T) {
	sign1 := GenerateSign("a", "b", "c", "d")
	sign2 := GenerateSign("a", "b", "c", "d")
	assert.Equal(t, sign1, sign2, "相同输入应产生相同签名")
}

func TestGenerateSignDifferentInputs(t *testing.T) {
	sign1 := GenerateSign("a", "b", "c", "d")
	sign2 := GenerateSign("a", "b", "c", "e")
	assert.NotEqual(t, sign1, sign2, "不同输入应产生不同签名")
}

func TestGetCurrentTimestamp(t *testing.T) {
	ts := GetCurrentTimestamp()
	assert.Len(t, ts, 14, "时间戳长度应为14")
}

func TestGenerateNonceStr(t *testing.T) {
	nonce := GenerateNonceStr(32)
	assert.Len(t, nonce, 32, "随机字符串长度应为32")

	for _, c := range nonce {
		assert.True(t, (c >= 'a' && c <= 'z') || (c >= 'A' && c <= 'Z') || (c >= '0' && c <= '9'),
			"随机字符串应只包含字母和数字")
	}
}

func TestGenerateNonceStrLengths(t *testing.T) {
	for _, length := range []int{16, 32, 64} {
		nonce := GenerateNonceStr(length)
		assert.Len(t, nonce, length, "随机字符串长度应匹配")
	}
}

func TestGenerateNonceStrUnique(t *testing.T) {
	set := make(map[string]bool)
	for i := 0; i < 100; i++ {
		nonce := GenerateNonceStr(16)
		assert.False(t, set[nonce], "每次生成的随机字符串应不同")
		set[nonce] = true
	}
}
