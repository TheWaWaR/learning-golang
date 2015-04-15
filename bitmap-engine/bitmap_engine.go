

package main;

import (
	"os"
	"log"
	"io"
	"bufio"
	"sort"
	"flag"
	"bytes"
	"encoding/binary"
	_ "time"
)

var (
	BLOCK_LENGTH = uint32(64)
	EXPANED_STEP = 64
	SET_OP_INTERSECTION = uint32(1)
	SET_OP_UNION        = uint32(2)
	ALL_BITS = [64]uint64{}

	// Bitmap files
	fn_bitmapA	= flag.String("b1", "", "First bitmap file")
	fn_bitmapB	= flag.String("b2", "", "Second bitmap file")
	fn_inputBitmap  = flag.String("iB", "", "input bitmap file")
	fn_outputBitmap = flag.String("oB", "", "Output bitmap file")
	// Ids files
	fn_idsA		= flag.String("i1", "", "First ids file(binary)")
	fn_idsB		= flag.String("i2", "", "Second ids file(binary)")
	fn_inputIds	= flag.String("iI", "", "input ids file(binary)")
	fn_outputIds	= flag.String("oI", "", "Output ids file(binary)")
	// Other args
	saveBitmap	= flag.Bool("save", true, "If automatic save bitmap to file")
	operation	= flag.String("op", "intersection", "Set operation type: [intersection, union]")
	debug		= flag.Bool("debug", false, "Debug mode")
)

func init() {
	current := uint64(1)
	for i := int(BLOCK_LENGTH-1); i >= 0; i-- {
		ALL_BITS[i] = current
		current = current << 1
	}
	// for i, value := range ALL_BITS { log.Printf("idx=%3d, value=%064b", i, value) }
	flag.Parse()
}

type BitMap struct {
	blocks []*BitMapBlock
	size uint32
	entropy uint32
}

type BitMapBlock struct {
	value uint64
	entropy uint32
}

type uints32 []uint32
// Len returns the length of the uints32 array.
func (x uints32) Len() int { return len(x) }
// Less returns true if element i is less than element j.
func (x uints32) Less(i, j int) bool { return x[i] < x[j] }
// Swap exchanges elements i and j.
func (x uints32) Swap(i, j int) { x[i], x[j] = x[j], x[i] }

/**
 * Common functions
 */
func howManyBlocks(ids []uint32, idsLength uint32) uint32 {
	lastId := ids[idsLength-1]
	nBlock := lastId / BLOCK_LENGTH
	if lastId % BLOCK_LENGTH > 0 {
		nBlock += 1
	}
	// log.Printf("last id: %d, n-block: %d", lastId, nBlock)
	return nBlock
}

func maxUint32(a uint32, b uint32) uint32 {
	if a > b { return a }
	return b
}

func loadIds(filename string) uints32 {
	f, err := os.Open(filename)
	if err != nil {
		panic(err)
	}
	defer f.Close()
	stats, statsErr := f.Stat()
	if statsErr != nil {
		panic(statsErr)
	}
	wordLength := int64(4)
	var size int64 = stats.Size()
	var idsLength = size / wordLength
	log.Printf("Load ids, file-size: %d", size)
	data := make([]byte, size)
	bufr := bufio.NewReader(f)
	_,err = bufr.Read(data)

	ids := make([]uint32, idsLength)
	var singleId uint32
	for startIdx := int64(0); startIdx < size; startIdx+=wordLength {
		part := bytes.NewReader(data[startIdx:startIdx+wordLength])
		binary.Read(part, binary.LittleEndian, &singleId)
		// log.Printf("single-id: %v, %d", part, singleId)
		ids[startIdx/wordLength] = uint32(singleId)
	}
	log.Printf("Loaded ids(%d)", len(ids))
	return ids
}


/* ============================================================================
 * Id array stuff
 * ==========================================================================*/
func (ids uints32) dump(filename string) {
	log.Printf("Dump ids to: %s", filename)
	f, err := os.OpenFile(filename, os.O_WRONLY | os.O_CREATE, 0644)
	if err != nil {
		panic(err)
	}
	defer f.Close()
	bufw := bufio.NewWriter(f)
	idsLength := len(ids)
	buf := new(bytes.Buffer)
	for i := 0; i < idsLength; i++ {
		binary.Write(buf, binary.LittleEndian, ids[i])
		buf.WriteTo(bufw)
	}
	bufw.Flush()
	log.Printf("Dumped ids(%d)", len(ids))
}

/* ============================================================================
 * BitMap stuff
 * ==========================================================================*/
func loadBitMap(filename string) *BitMap {
	f, err := os.Open(filename)
	if err != nil {
		panic(err)
	}
	defer f.Close()
	stats, statsErr := f.Stat()
	if statsErr != nil {
		panic(statsErr)
	}

	wordLength := uint64(8)
	size := uint64(stats.Size())
	var bitmapSize = uint32(uint64(size) / wordLength)
	log.Printf("bitmap file-size: %d", size)
	data := make([]byte, size)
	bufr := bufio.NewReader(f)
	_,err = bufr.Read(data)

	bitmap := &BitMap{
		blocks: make([]*BitMapBlock, bitmapSize),
		size: bitmapSize,
		entropy: 0,
	}
	var value uint64
	for startIdx := uint64(0); startIdx < size; startIdx += wordLength {
		part := bytes.NewReader(data[startIdx:startIdx+wordLength])
		binary.Read(part, binary.LittleEndian, &value)
		// log.Printf("single-id: %v, %d", part, singleId)
		block := NewBlock(value)
		bitmap.blocks[startIdx/wordLength] = block
		bitmap.entropy += block.entropy
	}
	return bitmap
}

func (bm *BitMap) writeTo(writer io.Writer) {
	for idx := uint32(0); idx < bm.size; idx++ {
		block := bm.blocks[idx]
		block.writeTo(writer)
	}
}

func (bm *BitMap) dump(filename string) {
	f, err := os.OpenFile(filename, os.O_WRONLY | os.O_CREATE, 0644)
	if err != nil {
		panic(err)
	}
	defer f.Close()
	bufw := bufio.NewWriter(f)
	bm.writeTo(bufw)
	bufw.Flush()
}

func buildBitMap(ids uints32) *BitMap {
	idsLength := uint32(len(ids))
	if (idsLength < 1) {
		log.Printf("Empty ids!")
		return nil
	}
	sort.Sort(ids)
	bitmapLength := howManyBlocks(ids, idsLength)
	bitmap := &BitMap{
		blocks: make([]*BitMapBlock, bitmapLength),
		size: bitmapLength,
		entropy: uint32(0),
	}

	startIdx := uint32(0)
	for blockIdx := uint32(0); blockIdx < bitmap.size; blockIdx++ {
		tmpIdx, block := buildBlock(blockIdx, startIdx, ids)
		startIdx = tmpIdx
		// log.Printf("block-idx: %d, block: %v", blockIdx, block)
		bitmap.blocks[blockIdx] = block
		bitmap.entropy += block.entropy
	}
	return bitmap
}

func (bm *BitMap) getIds() uints32 {
	allIds := make([]uint32, bm.entropy)
	idx := 0
	for i := uint32(0); i < bm.size; i++ {
		blockIds := bm.blocks[i].getIds()
		if blockIds != nil {
			base := i * BLOCK_LENGTH
			for _, id := range blockIds {
				allIds[idx] = base + id
				idx++
			}
		}
	}
	return allIds
}

func (bm *BitMap) trimTail() {
	for idx := int(bm.size-1); idx >= 0; idx-- {
		if bm.blocks[idx].value == 0 {
			bm.size--
		} else {
			break
		}
	}
}

func (bm *BitMap) expandToSize(newSize uint32) {
	newBlocks := make([]*BitMapBlock, newSize)
	newEntropy := uint32(0)
	for i := uint32(0); i < newSize; i++ {
		block := bm.getBlock(i)
		newBlocks[i] = block
		newEntropy += block.entropy
	}
	bm.blocks = newBlocks
	bm.size = newSize
	bm.entropy = newEntropy
}

func (bm *BitMap) show() {
	// log.Printf("@: BitMap = %v", bm)
	log.Printf("@ Bitmap for: %v", bm.getIds())
	log.Printf("block-length:%d, size=%d, entropy=%d",
		len(bm.blocks), bm.size, bm.entropy)
	for idx := uint32(0); idx < bm.size; idx++ {
		log.Printf("[BLOCK] %3d: %064b", idx, bm.blocks[idx].value)
	}
	log.Printf("----------------------------------------\n\n")
}

func (bm *BitMap) operate(otherBitMap *BitMap, op uint32) *BitMap {
	newSize := maxUint32(bm.size, otherBitMap.size)
	newBitMap := &BitMap{
		blocks: make([]*BitMapBlock, newSize),
		size: newSize,
		entropy: uint32(0),
	}
	var block *BitMapBlock
	for idx := uint32(0); idx < newSize; idx++ {
		switch (op) {
		case SET_OP_INTERSECTION:
			block = bm.getBlock(idx).intersection(otherBitMap.getBlock(idx))
		case SET_OP_UNION:
			block = bm.getBlock(idx).union(otherBitMap.getBlock(idx))
		}
		newBitMap.blocks[idx] = block
		// log.Printf("newBitMap.entropy=%d, block.entropy=%d",
		// 	newBitMap.entropy, block.entropy)
		newBitMap.entropy += block.entropy
	}
	newBitMap.trimTail()
	return newBitMap
}

func (bm *BitMap) intersection(otherBitMap *BitMap) *BitMap {
	return bm.operate(otherBitMap, SET_OP_INTERSECTION)
}

func (bm *BitMap) union(otherBitMap *BitMap) *BitMap {
	return bm.operate(otherBitMap, SET_OP_UNION)
}

func (bm *BitMap) getBlockDefault(idx uint32, defaultValue uint64) *BitMapBlock {
	var block *BitMapBlock
	if (idx < bm.size) {
		block = bm.blocks[idx]
	} else {
		block = &BitMapBlock{defaultValue, 0}
	}
	return block
}

func (bm *BitMap) getBlock(idx uint32) *BitMapBlock {
	return bm.getBlockDefault(idx, 0)
}

func (bm *BitMap) has(id uint32) bool {
	idx := id - 1
	blockIdx := idx / BLOCK_LENGTH
	block := bm.getBlock(blockIdx)
	return block.has(idx % BLOCK_LENGTH)
}

func (bm *BitMap) add(id uint32) {
	if bm.has(id) { return }

	idx := id - 1
	blockIdx := idx / BLOCK_LENGTH
	if blockIdx >= uint32(len(bm.blocks)) {
		// Expand blocks
		bm.expandToSize(blockIdx + 1)
	}
	if blockIdx >= bm.size {
		bm.size = blockIdx+1
	}
	idMod := idx % BLOCK_LENGTH
	bm.blocks[blockIdx].add(idMod)
	bm.entropy ++
}

func (bm *BitMap) remove(id uint32) {
	idx := id - 1
	blockIdx := idx / BLOCK_LENGTH
	idMod := idx % BLOCK_LENGTH
	block := bm.blocks[blockIdx]
	block.remove(idMod)
	if block.entropy == 0 {
		bm.trimTail()
	}
}

/* ============================================================================
 * Block stuff
 * ==========================================================================*/
func buildBlock(blockIdx uint32, startIdx uint32, ids uints32) (
	idx uint32, block *BitMapBlock) {

	startValue := blockIdx * BLOCK_LENGTH
	endValue := startValue + BLOCK_LENGTH
	idIdx := startIdx
	value := ids[idIdx]
	// log.Printf("startValue: %d, endValue: %d, startIdx: %d, value: %d",
	// 	startValue, endValue, startIdx, value)
	rv := uint64(0)
	entropy := uint32(0)
	for ; startValue < value && value <= endValue;  {
		modVal := value % BLOCK_LENGTH
		val := uint64(1)
		if modVal != 0 {
			val = val << (BLOCK_LENGTH - modVal)
		}
		if rv & val == 0 {
			entropy ++
		}
		rv |= val
		// log.Printf("value: %4d, val: %064b, rv: %064b", value, val, rv)
		if idIdx+1 < uint32(len(ids)) {
			idIdx++
			value = ids[idIdx]
		} else {
			break
		}
	}
	return idIdx, &BitMapBlock{rv, entropy}
}

func NewBlock(value uint64) *BitMapBlock {
	block := &BitMapBlock {
		value: value,
		entropy: 0,
	}
	block.updateEntropy()
	return block
}

func (block *BitMapBlock) intersection(otherBlock *BitMapBlock) *BitMapBlock {
	newBlock := &BitMapBlock{
		value: block.value & otherBlock.value,
		entropy: 0,
	}
	newBlock.updateEntropy()
	return newBlock
}

func (block *BitMapBlock) union(otherBlock *BitMapBlock) *BitMapBlock {
	newBlock := &BitMapBlock{
		value: block.value | otherBlock.value,
		entropy: 0,
	}
	newBlock.updateEntropy()
	return newBlock
}

func (block *BitMapBlock) updateEntropy() {
	entropy := uint32(0)
	value := block.value
	if value != uint64(0) {
		for _, bit := range ALL_BITS {
			if (value & bit) > 0 {
				entropy ++
			}
		}
	}
	// log.Printf("old-entropy=%d, new-entropy=%d", block.entropy, entropy)
	block.entropy = entropy
}

func (block *BitMapBlock) writeTo(writer io.Writer) {
	buf := new(bytes.Buffer)
	binary.Write(buf, binary.LittleEndian, block.value)
	buf.WriteTo(writer)
}

func (block *BitMapBlock) getIds() uints32 {
	value := block.value
	var ids uints32 = nil
	if value != uint64(0) {
		ids = make(uints32, block.entropy)
		idx := 0
		for i, bit := range ALL_BITS {
			if (value & bit) > 0 {
				ids[idx] = uint32(i+1)
				idx++
			}
		}
	}
	return ids
}

func (block *BitMapBlock) has(idMod uint32) bool {
	mask := ALL_BITS[idMod]
	return (block.value & mask > 0)
}
func (block *BitMapBlock) add(idMod uint32) {
	mask := ALL_BITS[idMod]
	block.value = block.value | mask
	block.updateEntropy()
}
func (block *BitMapBlock) remove(idMod uint32) {
	mask := ALL_BITS[idMod]
	block.value = block.value & (^mask)
	block.updateEntropy()
}


/* ============================================================================
 *  Test functions
 * ==========================================================================*/
func makeBitMaps(idsFilenames []string, bitmapFilenames []string) []*BitMap {
	bitmaps := make([]*BitMap, len(idsFilenames))
	for idx, idsFilename := range idsFilenames {
		ids := loadIds(idsFilename)
		// log.Printf("ids: %v", ids)
		bitmap := buildBitMap(ids)
		bitmaps[idx] = bitmap
	}
	for idx, bitmap := range bitmaps {
		bitmap.show()
		bitmap.dump(bitmapFilenames[idx])
	}
	return bitmaps
}

func testIntersection(bitmaps []*BitMap, fileIds string, fileBitmap string) {
	log.Printf(">>> Test intersection")
	log.Printf("--------------------")
	bitmapIntersection := bitmaps[0].intersection(bitmaps[1])
	bitmapIntersection.show()
	bitmapIntersection.getIds().dump(fileIds)
	bitmapIntersection.dump(fileBitmap)
}

func testUnion(bitmaps []*BitMap, fileIds string, fileBitmap string) {
	log.Printf(">>> Test union")
	log.Printf("--------------------")
	bitmapUnion := bitmaps[0].union(bitmaps[1])
	bitmapUnion.show()
	bitmapUnion.getIds().dump(fileIds)
	bitmapUnion.dump(fileBitmap)
}

func testLoadBitmaps(bitmapFilenames []string) {
	for _, filename := range bitmapFilenames {
		bitmap := loadBitMap(filename)
		log.Printf("filename=%s, bitmap=%v", filename, bitmap)
		bitmap.show()
	}
}

func tests() {
	idsFilenames := []string{"/tmp/days-1.ids", "/tmp/days-2.ids"}
	bitmapFilenames := []string{"/tmp/days-1.bitmap", "/tmp/days-2.bitmap"}
	bitmaps := makeBitMaps(idsFilenames, bitmapFilenames)
	testIntersection(bitmaps, "/tmp/intersection.ids", "/tmp/intersection.bitmap")
	testUnion(bitmaps, "/tmp/union.ids", "/tmp/union.bitmap")
	testLoadBitmaps(bitmapFilenames)
}

func idsToBitmap(fn_inputIds string, fn_outputBitmap string) {
	buildBitMap(loadIds(fn_inputIds)).dump(fn_outputBitmap)
}

func bitmapToIds(fn_inputBitmap string, fn_outputIds string) {
	loadBitMap(fn_inputBitmap).getIds().dump(fn_outputIds)
}

func main() {
	log.Printf("Args: bitmapA=(%v), bitmapB=(%v), inputBitmap=(%v), outputBitmap=(%v)",
		*fn_bitmapA, *fn_bitmapB, *fn_inputBitmap, *fn_outputBitmap)
	log.Printf("Args: idsA=(%v), idsB=(%v), inputIds=(%v), outputIds=(%v)",
		*fn_idsA, *fn_idsB, *fn_inputIds, *fn_outputIds)
	log.Printf("Args: saveBitmap=(%v), operation=(%v), debug=(%v)", *saveBitmap, *operation, *debug)

	// Format transform
	if len(*fn_inputIds) > 0 && len(*fn_outputBitmap) > 0 {
		log.Printf("Transform ids=%v (to) bitmap=%v", *fn_inputIds, *fn_outputBitmap)
		idsToBitmap(*fn_inputIds, *fn_outputBitmap)
		return
	}
	if len(*fn_inputBitmap) > 0 && len(*fn_outputIds) > 0 {
		log.Printf("Transform bitmap=%v (to) ids=%v", *fn_inputBitmap, *fn_outputIds)
		bitmapToIds(*fn_inputBitmap, *fn_outputIds)
		return
	}


	fn_bA := *fn_bitmapA
	if len(*fn_bitmapA) == 0 {
		fn_bA = *fn_idsA + ".bitmap"
	}
	fn_bB := *fn_bitmapB
	if len(*fn_bitmapB) == 0 {
		fn_bB = *fn_idsB + ".bitmap"
	}

	var err error
	_, err = os.Stat(fn_bA)
	bitmapAexists := err == nil
	_, err = os.Stat(fn_bB)
	bitmapBexists := err == nil
	_, err = os.Stat(*fn_idsA)
	idsAexists := err == nil
	_, err = os.Stat(*fn_idsB)
	idsBexists := err == nil
	if (!bitmapAexists && !idsAexists) || (!bitmapBexists && !idsBexists) {
		log.Fatalf("Input A or input B not given !!!")
	}
	if (len(*fn_outputIds) == 0 && len(*fn_outputBitmap) == 0) {
		log.Fatalf("Output file not given !!!")
	}

	var bitmapA, bitmapB, outputBitmap *BitMap
	// Load bitmaps

	if bitmapAexists {
		// bitmapA exists
		log.Printf("Load bitmapA from: %s", fn_bA)
		bitmapA = loadBitMap(fn_bA)
		log.Printf("Loaded bitmapB")
	} else if idsAexists {
		log.Printf("Load idsA from: %s", *fn_idsA)
		idsA := loadIds(*fn_idsA)
		log.Printf("Loaded idsA")
		bitmapA = buildBitMap(idsA)
		if *saveBitmap {
			log.Printf("Dump bitmap to: %v", fn_bA)
			bitmapA.dump(fn_bA)
			log.Printf("Finished!")
		}
	}

	if bitmapBexists {
		// bitmapA exists
		log.Printf("Load bitmapB from: %s", fn_bB)
		bitmapB = loadBitMap(fn_bB)
		log.Printf("Loaded bitmapB")
	} else if idsBexists {
		log.Printf("Load idsB from: %s", *fn_idsB)
		idsB := loadIds(*fn_idsB)
		log.Printf("Loaded idsB")
		bitmapB = buildBitMap(idsB)
		if *saveBitmap {
			log.Printf("Dump bitmap to: %v", fn_bB)
			bitmapB.dump(fn_bB)
			log.Printf("Finished!")
		}
	}

	if (*debug) {
		bitmapA.show()
		bitmapB.show()
	}

	// Set operations
	log.Printf("Starting... operation=%s", *operation)
	switch (*operation) {
	case "intersection":
		outputBitmap = bitmapA.intersection(bitmapB)
	case "union":
		outputBitmap = bitmapA.union(bitmapB)
	}
	log.Printf("Finished operation=%s", *operation)

	if (*debug) {
		outputBitmap.show()
	}
	// Dump ids
	if len(*fn_outputIds) > 0 {
		outputIds := outputBitmap.getIds()
		if (*debug) {
			log.Printf("outputIds: %v", outputIds)
		}
		log.Printf("Dump outputIds: count=%d", len(outputIds))
		outputIds.dump(*fn_outputIds)
	}
	// Dump bitmaps
	if len(*fn_outputBitmap) > 0 {
		log.Printf("Dump outputBitmap.entropy=%d", outputBitmap.entropy)
		outputBitmap.dump(*fn_outputBitmap)
	}
}
