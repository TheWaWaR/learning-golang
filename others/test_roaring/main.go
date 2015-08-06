package main

import (
	"log"
	"github.com/tgruben/roaring"
)


func main() {
	rb1 := roaring.NewRoaringBitmap()
	for i := 0; i < 5000000; i++ {
		rb1.Add(i)
	}
	log.Printf("rb1.GetSerializedSizeInBytes(): %d", rb1.GetSerializedSizeInBytes())
	log.Printf("rb1.GetSizeInBytes(): %d", rb1.GetSizeInBytes())

	log.Printf("=================================")
	for n := 2; n <= 100; n++ {
		rbn := roaring.NewRoaringBitmap()
		for i := 0; i < 10000000; i++ {
			if i % n == 0 {
				rbn.Add(i)
			}
		}
		log.Printf("rb%d.GetSerializedSizeInBytes(): %d", n, rbn.GetSerializedSizeInBytes())
		log.Printf("rb%d.GetSizeInBytes(): %d", n, rbn.GetSizeInBytes())
		log.Printf("--------------------")
	}
}
