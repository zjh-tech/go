package main

import (
	"reflect"
	"time"

	"github.com/zjh-tech/go-frame/engine/elog"
)

func main() {
	logger := elog.NewLogger("./testreflect", 0)
	logger.Init()
	ELog.SetLogger(logger)

	//TestType()
	//TestGetValue()
	//TestSetValue()
	TestStruct()

	for {
		time.Sleep(1 * time.Second)
	}
}

//拿类型
func reflectType(x interface{}) {
	obj := reflect.TypeOf(x)
	ELog.InfoAf("Obj=%v Type=%v Kind=%v", obj, obj.Name(), obj.Kind())
}

type Cat struct {
}

type Dog struct {
}

func TestType() {
	var a float32 = 1.234
	reflectType(a)
	var b int8 = 10
	reflectType(b)
	var c Cat
	reflectType(c)
	var d Dog
	reflectType(d)
	//Go 数组,切片,Map,指针等类型的变量,他们的.Name()都是返回空
	//slice
	var e []int
	reflectType(e)
	var f []string
	reflectType(f)
}

//拿值
func reflectGetValue(x interface{}) {
	v := reflect.ValueOf(x)
	ELog.InfoAf("%v %T", v, v)
	k := v.Kind()
	switch k {
	case reflect.Float32:
		ret := float32(v.Float())
		ELog.InfoAf("%v,%T", ret, ret)
	case reflect.Int32:
		ret := int32(v.Int())
		ELog.InfoAf("%v,%T", ret, ret)
	}
}

func TestGetValue() {
	var aa int32 = 100
	reflectGetValue(aa)

	var bb float32 = 1.234
	reflectGetValue(bb)
}

//设置值
func reflectSetValue(x interface{}) {
	v := reflect.ValueOf(x)
	//Elem 根据指针取到对应的值
	k := v.Elem().Kind()
	switch k {
	case reflect.Float32:
		v.Elem().SetFloat(3.21)
	case reflect.Int32:
		v.Elem().SetInt(100)
	}
}

func TestSetValue() {
	var aaa int32 = 10
	reflectSetValue(&aaa)
	ELog.InfoA(aaa)
}

type student struct {
	Name  string `json:"name"`
	Score int    `json:"score"`
}

func (s *student) Study() string {
	msg := "好好学习;天天向上"
	ELog.InfoA(msg)
	return msg
}

func (s *student) Sleep() string {
	msg := "好好睡觉;快快长大"
	ELog.InfoA(msg)
	return msg
}

func TestStruct() {
	stu := student{
		Name:  "小王子",
		Score: 90,
	}

	t := reflect.TypeOf(stu)
	ELog.InfoAf("Type=%v Kind=%v", t.Name(), t.Kind())

	//遍历结构体所有字段
	for i := 0; i < t.NumField(); i++ {
		//根据结构体字段的索引去去字段
		fieldObj := t.Field(i)
		ELog.InfoAf("FiledObj=%+v", fieldObj)
	}

	//根据名字去取结构体中的字段
	filedObj2, ok := t.FieldByName("Score")
	if ok {
		ELog.InfoAf("FiledObj2=%+v", filedObj2)
	}

	v := reflect.ValueOf(stu)

	for i := 0; i < v.NumMethod(); i++ {
		ELog.InfoAf("MethodType=%v Name=%v", v.Method(i).Type(), t.Method(i).Name)

		//通过发射调用方法传递的参数必须是[]reflect.Value类型
		var args = []reflect.Value{}
		v.Method(i).Call(args) //调用方法
	}
}
