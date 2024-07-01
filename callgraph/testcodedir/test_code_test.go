package testcodedir

import "fmt"

func GlobalFunc() {
}

const globalStr = "Hello, World!"

func mainx() {
	fmt.Println(globalStr)
	abc := &AbcValueRecv{}
	CallInterfaceFunc(abc)
}

type AbcValueRecv struct {
	Name string
}

func (abc AbcValueRecv) MemberFunc() {
	GlobalFunc()
}

type AbcPointerRecv struct {
}

func (abc *AbcPointerRecv) MemberFunc() {

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
