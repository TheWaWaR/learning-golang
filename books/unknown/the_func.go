
package main

import (
)

func test_defer() {
	for i := 0; i < 5; i++ {
		defer func() { println("closure:", i) }() // 可读写UpValue的闭包
		defer print("index:", i, ",") // i 是值拷贝的？
	}
}

func vararg(arg ...interface{}) {
	a := func() {
		print("a")
	}
	a()
	
	var b func(int)
	b = func(i int) {
		print("b:", i)
	}
	b(1)
}


func main() {
	test_defer()
}


