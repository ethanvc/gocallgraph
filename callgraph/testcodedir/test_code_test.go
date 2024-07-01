package testcodedir

import "fmt"

func GlobalFunc() {
}

const globalStr = "Hello, World!"

func mainx() {
	fmt.Println(globalStr)
	abc := &Abc{}
	CallInterfaceFunc(abc)
}

type Abc struct {
	Name string
}

func (abc Abc) MemberFunc() {
	GlobalFunc()
}

func CallInterfaceFunc(abc AbcInterface) {
	abc.MemberFunc()
}

func CallGlobalFunc(abc AbcInterface) {
	GlobalFunc()
}

type AbcInterface interface {
	MemberFunc()
}
