package main


// Ref>> http://golang.org/src/pkg/image/png/reader_test.go
import (
	"os"
	"fmt"
	"image"
	"sort"
//	"image/color"
	"image/png"
)

func readPNG(filename string) (image.Image, error) {
	f, err := os.Open(filename)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	return png.Decode(f)
}

type ColorCounter struct {
	val uint32
	cnt int
}
type ByCount []ColorCounter
func (a ByCount) Len() int { return len(a) }
func (a ByCount) Swap(i int, j int) { a[i], a[j] = a[j], a[i] }
func (a ByCount) Less(i int, j int) bool {
	if (a[i].cnt == a[j].cnt) {
		return a[i].val < a[j].val
	} else {
		return a[i].cnt < a[j].cnt
	}
}


func main() {
	defer func() {
		if r := recover(); r != nil {
			println(fmt.Sprintf("%q", r))
		}
	}()
	
	png_fn := "test.png"
	img, err := readPNG(png_fn)
	if err != nil {}

	rect := img.Bounds()
	rb := rect.Max
	fmt.Println("Rect: %x", rect)
	println("--------------------")

	colors := make(map[uint32]int)
	
	for x := 0; x < rb.X; x++ {
		for y := 0; y < rb.Y; y++ {
			c := img.At(x, y)
			r, g, b, _ := c.RGBA()
			
			val := ((r>>8)<<16) | ((g>>8)<<8) | (b>>8)
			// println(fmt.Sprintf("r:%x, g:%x, b:%x, val:%x", r, g, b, val))
			cnt, ok := colors[val]
			if ok {
				cnt += 1
			} else {
				cnt = 1
			}
			colors[val] = cnt
			
			// println(fmt.Sprintf("Color[%d, %d]: %x", x, y, c))
		}
	}

	ccs := make([]ColorCounter, len(colors))
	idx := 0
	total := 0
	for val, cnt := range colors {
		ccs[idx] = ColorCounter{val: val, cnt: cnt}
		// println(fmt.Sprintf("Color[%d]: %d", val, cnt))
		total += cnt
		idx += 1
	}

	sort.Sort(ByCount(ccs))
	println(fmt.Sprintf("ColorCounter.length: %d", len(ccs)))
	println("--------------------")

	for _, cc := range ccs {
		println(fmt.Sprintf("color: %7x  -->  count: %6d", cc.val, cc.cnt))
		// fmt.Sprintf("val: %12d, cnt: %6d", cc.val, cc.cnt)
	}
	pixelCnt := rb.X * rb.Y
	println("----------------------------------------")
	last2_cnt := ccs[idx-1].cnt + ccs[idx-2].cnt
	println(fmt.Sprintf(" > total==pixelCnt -> %t \n > maxCnt=%d, pixelCnt=%d \n >> prec=%f",
		total==pixelCnt,
		ccs[len(ccs)-1].cnt,
		pixelCnt,
		float32(last2_cnt) / float32(pixelCnt)))
}
