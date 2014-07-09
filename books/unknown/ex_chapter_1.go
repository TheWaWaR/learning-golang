
package main

import (
	"fmt"
	"os"
	"bufio"
	"unicode/utf8"
)


func q1() {
	// For-loop
	for i := 0; i < 10; i++ {
		fmt.Printf("%d", i)
	}
	println("\n---")
	
	j := 0			// init
LOOP:
	if j < 10 {
		// Body
		fmt.Printf("%d", j)
		// Last
		j++
		goto LOOP
	}
	println("\n---")
	
	arr := [...]int{11, 21, 31}
	for idx, val := range arr {
		println(idx, val)
	}
	println("=======")
}

func q2() {
	// FizzBuzz
	for i := 1; i <= 100; i++ {
		/*
		switch {
		case i%3 == 0 && i%5 == 0:
			println(".Fizz-Buzz")
		case i%3 == 0: 
			print(".")
			println("Fizz")
		case i%5 == 0: 
			print(".")
			println("Buzz")
		}  */
		
		if i%5 == 0 && i%3 == 0 {
			println("Fizz-Buzz")
		} else if i%3 == 0 {
			println("Fizz")
		} else if i%5 == 0 {
			println("Buzz")
		}
	}
	println("=======")
}

func q3() {
	for i := 1; i <= 10; i++ {
		for j := 0; j < i; j++ { print("A") }
		println("")
	}
	println("----")
	bio := bufio.NewReader(os.Stdin)
	for {
		line, err := bio.ReadString('\n')
		if err != nil {
			println("EOF:", err)
			break
		}
		line = line[0:len(line)-1]
		c := []rune(line)
		sum := 0
		for _, r := range c {
			sum += utf8.RuneLen(r)
		}
		println(line, "-->", len(c), "==>", sum)
		println("----")
		

		cc := make([]rune, len(c))
		copy(cc, c)
		for head, tail := 0, len(c)-1; head < tail; head, tail = head+1, tail-1 {
			cc[head], cc[tail] = cc[tail], cc[head]
		}
		println("Reversed:", string(cc))
		
		if len(c) >= 7 {
			println("----")
			for idx, nc := range []rune("abc") { c[4+idx] = nc }
			println("Replaced:", string(c))
		}
		println(">>--NEW--<<")
	}
	println("=======")
}

func q4() {
	arr := [...]float64{1.1, 2.2, 3, 4, -5.3}
	sl := make([]float64, 5)
	sl = arr[:]
	sum := float64(0)
	for _, val := range sl {
		sum += val
	}
	println(sum)
	println("=======")
}


func main() {
	q1()
	q2()
	q3()
	q4()
}
