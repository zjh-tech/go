package convert

import (
	"math/rand"
	"testing"
)

func Test_Uint64(t *testing.T) {
	fromValue := rand.Uint64()
	str := Uint64ToStr(fromValue)
	toValue, err := StrToUint64(str)
	if err != nil || fromValue != toValue {
		t.Fail()
	}
}

func Test_Uint32(t *testing.T) {
	fromValue := rand.Uint32()
	str := Uint32ToStr(fromValue)
	toValue, err := StrToUint32(str)
	if err != nil || fromValue != toValue {
		t.Fail()
	}
}

func Test_Int64(t *testing.T) {
	fromValue := rand.Int63()
	str := Int64ToStr(fromValue)
	toValue, err := StrToInt64(str)
	if err != nil || fromValue != toValue {
		t.Fail()
	}
}

func Test_Int32(t *testing.T) {
	fromValue := rand.Int31()
	str := Int32ToStr(fromValue)
	toValue, err := StrToInt32(str)
	if err != nil || fromValue != toValue {
		t.Fail()
	}
}

func Test_Int(t *testing.T) {
	fromValue := rand.Int()
	str := IntToStr(fromValue)
	toValue, err := StrToInt(str)
	if err != nil || fromValue != toValue {
		t.Fail()
	}
}
