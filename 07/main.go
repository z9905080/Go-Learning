//go:generate stringer -type=Pill
package main

import (
	"fmt"
	"reflect"
	ConCurrentMap "userLibarary/ConCurrentMap"
)

func main() {

	cmap := ConCurrentMap.NewConcurrentMap(reflect.TypeOf(int(0)), reflect.TypeOf(string(0)))
	fmt.Println(cmap)
	cmap.Put(0, "456")
	//data, err := cmap.Put(113, 12)
	fmt.Println(cmap)

}
