package main

type MyStructInterFace interface {
	Set(string)
}

func NewMyInstance(mi MyStructInterFace, str string) {
	mi.Set(str)
}

type MyStruct struct {
	value string
}

func (m *MyStruct) Set(newVar string) {
	m.value = newVar
}

func (m MyStruct) Set2(newVar string) MyStruct {
	m.value = newVar
	return m
}

func main() {
	mStruct := MyStruct{}
	mStruct.Set("123")            //call by address
	mStruct = mStruct.Set2("456") //call by value 參數傳遞淺層複製的原因
	// temp := "123"
	// NewMyInstance(&mStruct, temp)
	//fmt.Println(mStruct)
}
