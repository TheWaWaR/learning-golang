
package main

import "fmt"


func testGoto() {
	i := 0
Here:
	println(i)
	i++
	if i < 10 { goto Here }
	println("THE END.")
	println("========")
}

func testLabelFor() {
	J: for j := 0; j < 5; j++ {
		for i := 0; i < 5; i++ {
			if j > 3 { break J }
			println("i: %d", i)
		}
		println("> j: %d", j)
	}
	println("========")
}

func testListArray() {
	strList := []string{"A", "B", "C", "DD", "EEE", "FFFF"}
	for idx, s := range strList {
		println(idx, s)
	}
	println("========")
}

func testSlice() {
	a := [...]int{ 1, 2, 3, 4, 5 }
	sl2 := a[2:4]
	sl3 := a[:4]
	println(len(sl2), cap(sl2), len(a))
	println(len(sl3), cap(sl3))
	println("========")
}

func testMap() {
	dict := map[string]string{
		"a": "A", "B": "BB",
		"c": "CCC", "D": "DDDD",
		"e": "EEEEE", "f": "FFFFFF",
	}
	println(dict["c"])
	delete(dict, "c")
	for k, v := range dict {
		println(k, v)
	}
	dict1 := make(map[string]int)
	dict2 := map[string]int{}
	println(dict1, dict2)
	println("========")
}


func main() {
	
	// 变量声明
	var ta int;
	var tb bool;
	ta = 15
	tb = false
	fmt.Printf("{%d, %d}\n", ta, tb)
	
	a := 15
	b := false
	fmt.Printf("{%d, %d}\n", a, b)

	// >> if 
	if a > 10 { fmt.Printf("Yes, a>10\n") }
	// >> goto
	testGoto()
	// >> for
	testLabelFor()
	// >> list aray
	testListArray()
	// >> slice
	testSlice()
	// >> map
	testMap()
	
	/*
	var (
		tx int
		ty bool
	)
	x, y, z := 1, 2, 3
	_, i := 3.3, 4.6
        */

	const (
		const_a = iota
		const_b = iota
	)
}
