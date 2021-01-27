package eutil

import (
	"bytes"
	"crypto/des"
	"crypto/md5"
	"encoding/base64"
	"encoding/hex"
	"errors"
)

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

//RC算法

//func test() {
//	src := "abcdefg"
//	md5_out := eutil.Md5([]byte(src))
//	fmt.Println(md5_out)
//	fmt.Println()
//
//	base64_out := eutil.EncryptBase64([]byte(src))
//	fmt.Println(base64_out)
//	base64_out1, _ := eutil.DecryptBase64(base64_out)
//	fmt.Println(string(base64_out1))
//	fmt.Println()
//
//	key := "1234567890"
//	des_out, _ := eutil.EncryptDes([]byte(src), key)
//	fmt.Println(base64.StdEncoding.EncodeToString(des_out))
//	des_out1, _ := eutil.DecryptDes(des_out, key)
//	fmt.Println(string(des_out1))
//	fmt.Println()
//
//	rc4obj, _ := rc4.NewCipher([]byte(key))
//	rc4_out := make([]byte, len(src))
//	rc4obj.XORKeyStream(rc4_out, []byte(src))
//	fmt.Println(base64.StdEncoding.EncodeToString(rc4_out))
//}
