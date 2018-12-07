package main

import "testing"

func BenchmarkSet(b *testing.B) {

	for i := 0; i < b.N; i++ {
		mStruct := MyStruct{}
		mStruct.Set("123")
	}
}
func BenchmarkSetWithInterface(b *testing.B) {

	for i := 0; i < b.N; i++ {
		mStruct := MyStruct{}
		NewMyInstance(&mStruct, "123")
	}
}

func BenchmarkSet2(b *testing.B) {

	for i := 0; i < b.N; i++ {
		mStruct := MyStruct{}
		mStruct = mStruct.Set2("456")
	}
}
