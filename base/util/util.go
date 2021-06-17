package util

import (
	"bytes"
	"crypto/des"
	"crypto/md5"
	"encoding/base64"
	"encoding/binary"
	"encoding/hex"
	"errors"
	"hash/fnv"
	"math/rand"
	"net"
	"os"
	"path/filepath"
	"strings"
	"time"
)

//time
func GetMillsecond() int64 {
	return time.Now().UnixNano() / 1e6
}

func GetSecond() uint32 {
	return uint32(time.Now().UnixNano() / 1e9)
}

//file
func GetExeDirectory() (string, error) {
	dir, err := filepath.Abs(filepath.Dir(os.Args[0]))
	if err != nil {
		return "", err
	}
	return strings.Replace(dir, "\\", "/", -1), nil
}

//rand  [0,n)
func GetRandom(min int, max int) int {
	if min > max {
		temp := min
		min = max
		max = temp
	}

	distance := max - min
	return rand.Intn(distance) + min
}

func GetRandomFloat64(min float64, max float64) float64 {
	if min > max {
		temp := min
		min = max
		max = temp
	}

	rdm := rand.Float64()
	return rdm*(max-min) + min
}

// 随机,M中取N,N<=M
func RandomNOfM(n int, m int) (idxs []int) {
	if n > m {
		return nil
	}

	idx := make([]int, m)
	idx[0] = -1 //索引0特殊处理
	for i := 0; i < n; i++ {
		r := i + rand.Intn(m-i)
		if idx[i] == 0 {
			idx[i] = i
		}
		if idx[r] == 0 {
			idx[r] = r
		}
		idx[i], idx[r] = idx[r], idx[i]
		if idx[i] == -1 {
			idx[i] = 0
		}
	}

	return idx[:n]
}

//洗牌
func Shuffle(s []int) {
	l := len(s)
	if l < 2 {
		return
	}

	for i := 0; i < l; i++ {
		r := rand.Intn(l)
		s[i], s[r] = s[r], s[i]
	}
}

func SliceRemoveByIndex(s []interface{}, i int) {
	s = append(s[:i], s[i+1:]...)
}

func SliceRemoveByElem(s []interface{}, value interface{}) {

}

//math.MaxUint32
func EnsureRange(value *int32, min int32, max int32) {
	if *value < min {
		*value = min
	}

	if *value > max {
		*value = max
	}
}

//config Split Char
const (
	OneSplitChar   string = "#"
	TwoSplitChar   string = "|"
	ThreeSplitChar string = "^"
)

func Uint32ToNetBytes(value uint32) []byte {
	buf := make([]byte, 4)
	binary.BigEndian.PutUint32(buf, value)
	return buf
}

func NetBytesToUint32(buf []byte) uint32 {
	return binary.BigEndian.Uint32(buf)
}

func GetLocalIp() (string, error) {
	addrs, err := net.InterfaceAddrs()
	if err != nil {
		return "", err
	}

	for _, address := range addrs {
		if ipnet, ok := address.(*net.IPNet); ok && !ipnet.IP.IsLoopback() {
			if ipnet.IP.To4() != nil {
				return ipnet.IP.String(), nil
			}
		}
	}

	return "", errors.New("Local Ip Not Find")
}

func GetTotalDayByMonth(year uint64, month uint64) (bool, uint64) {
	if month < 1 || month > 13 || year <= 0 {
		return false, uint64(0)
	}

	if month == 13 {
		month = 1
		year = year + 1
	}

	if month == 1 || month == 3 || month == 5 || month == 7 || month == 8 || month == 10 || month == 12 {
		return true, 31
	}

	if month == 2 {
		if IsLeapYear(year) {
			return true, 29
		} else {
			return true, 28
		}
	}

	return true, 30
}

func IsLeapYear(year uint64) bool {
	return (year%100 != 0 && year%4 == 0) || (year%400 == 0)
}

func Hash(str string) uint32 {
	hash := fnv.New32()
	hash.Write([]byte(str))
	return hash.Sum32()
}

func Hash64(str string) uint64 {
	hash := fnv.New64()
	hash.Write([]byte(str))
	return hash.Sum64()
}

//md5
func Md5(data []byte) string {
	has := md5.New()
	has.Write(data)
	return hex.EncodeToString(has.Sum(nil))
}

//base64
func EncryptBase64(data []byte) string {
	return base64.StdEncoding.EncodeToString(data)
}

func DecryptBase64(data string) ([]byte, error) {
	out, err := base64.StdEncoding.DecodeString(data)
	if err != nil {
		return nil, err
	}
	return out, nil
}

//DES算法 ECB Mode
func EncryptDes(data []byte, key string) ([]byte, error) {
	key_bytes := []byte(key)

	if len(key_bytes) > 8 {
		key_bytes = key_bytes[:8]
	}
	block, err := des.NewCipher(key_bytes)
	if err != nil {
		return nil, err
	}
	bs := block.BlockSize()
	data = pkcs5_padding(data, bs)
	if len(data)%bs != 0 {
		return nil, errors.New("EncryptDes Need a multiple of the blocksize")
	}
	out := make([]byte, len(data))
	dst := out
	for len(data) > 0 {
		block.Encrypt(dst, data[:bs])
		data = data[bs:]
		dst = dst[bs:]
	}
	return out, nil
}

func DecryptDes(data []byte, key string) ([]byte, error) {
	key_bytes := []byte(key)
	if len(key_bytes) > 8 {
		key_bytes = key_bytes[:8]
	}
	block, err1 := des.NewCipher(key_bytes)
	if err1 != nil {
		return nil, err1
	}
	bs := block.BlockSize()
	if len(data)%bs != 0 {
		return nil, errors.New("DecryptDES crypto/cipher: input not full blocks")
	}
	out := make([]byte, len(data))
	dst := out
	for len(data) > 0 {
		block.Decrypt(dst, data[:bs])
		data = data[bs:]
		dst = dst[bs:]
	}
	out = pkcs5_unpadding(out)
	return out, nil
}

func pkcs5_padding(ciphertext []byte, blockSize int) []byte {
	padding := blockSize - len(ciphertext)%blockSize
	padtext := bytes.Repeat([]byte{byte(padding)}, padding)
	return append(ciphertext, padtext...)
}

func pkcs5_unpadding(origData []byte) []byte {
	length := len(origData)
	unpadding := int(origData[length-1])
	return origData[:(length - unpadding)]
}
