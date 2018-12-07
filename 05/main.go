//go:generate stringer -type=Pill
package main

import (
	"fmt"
	"os"
	"runtime"
	"time"
)

type A interface {
	A()
}
type B interface {
	A
	B()
}
type NewT int

func (nt NewT) A() {
	fmt.Println("A")
}
func (nt NewT) B() {
	fmt.Println("B")
}
func createA(a A) {
	a.A()
}
func createB(b B) {
	b.A()
	b.B()
}

type Copyable interface {
	Copy() Value
}

type Value []int

func (v Value) Copy() Value { return v }

func main() {

	value := Value([]int{1, 2, 3, 4})

	var copy Copyable = value
	gg := copy.Copy()
	fmt.Printf("%p, %v\n", value, value)
	fmt.Printf("%p, %v\n", gg, gg)

	// t := []int{1, 2, 3, 4}
	// ss := make([]interface{}, len(t))
	// var value *interface{}
	// for i, v := range t {
	// 	ss[i] = v
	// 	if ss[i] == 4 {
	// 		*value = ss[i]
	// 	}
	// }
	// fmt.Println(ss)
	// fmt.Println(*value)
	// *value = 5
	// fmt.Println(ss)
	// fmt.Println(*value)

	var s NewT = 5
	createA(s)
	createB(s)

	fmt.Println(runtime.Version())
	f := createFile("foo1.txt")
	time.Sleep(time.Second * 1)
	defer closeFile(f)
	time.Sleep(time.Second * 1)
	writeFile(f)

}
func createFile(p string) *os.File {
	fmt.Println("creating")
	f, err := os.Create(p)
	defer println("Create defer!")
	if err != nil {
		panic(err)
	}
	return f
}
func writeFile(f *os.File) {
	defer println("writing defer!")
	fmt.Println("writing")
	fmt.Fprintln(f, "data")

}
func closeFile(f *os.File) {
	defer println("closing defer!")
	fmt.Println("closing")
	f.Close()
}
