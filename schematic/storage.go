package schematic

import (
	"fmt"
	"math"
	"math/bits"
)

const defaultBits = 1

type bitArray struct {
	data []int64

	//bitsPerEntry Number of bits an entry takes up
	bitsPerEntry int

	//maxEntryValue The maximum value for a single entry, Also works as a bitmask.
	//For example, if bitPerEntry is 5, maxEntryValue would be 31( (1 << bitPerEntry) -1)
	maxEntryValue int64

	//Number of entries in this array(blocks number)
	entrySize int
}

func NewEmptyBitArray(EntrySize int) *bitArray {
	b := &bitArray{
		data:          nil,
		bitsPerEntry:  defaultBits,
		maxEntryValue: (1 << defaultBits) - 1,
		entrySize:     EntrySize,
	}

	dataLen := int(math.Ceil(float64(defaultBits*EntrySize) / 64))

	b.data = make([]int64, dataLen)

	return b
}

func NewLitematicaBitArray(bits, EntrySize int, data []int64) *bitArray {
	b := &bitArray{
		data:          data,
		bitsPerEntry:  bits,
		maxEntryValue: (1 << bits) - 1,
		entrySize:     EntrySize,
	}

	dataLen := int(math.Ceil(float64(bits*EntrySize) / 64))

	if data != nil {
		if len(data) != dataLen {
			panic(fmt.Errorf("data length error %d, %d", len(data), dataLen))
		}
	} else {
		b.data = make([]int64, dataLen)
	}
	return b
}

func (b *bitArray) BitsPerEntry() int {
	return b.bitsPerEntry
}

func (b *bitArray) getBlock(index int64) int {
	return b.getAt(index)
}

func (b *bitArray) getAt(index int64) int {
	startOffset := index * int64(b.bitsPerEntry)
	startArrIndex := int(startOffset >> 6) // startOffset / 64
	endArrIndex := int(((index+1)*int64(b.bitsPerEntry) - 1) >> 6)
	startBitOffset := int(startOffset & 0x3F)

	if len(b.data) <= startArrIndex {
		return 0
	}
	if startArrIndex == endArrIndex {
		return int(int64(uint(b.data[startArrIndex])>>uint(startBitOffset)) & b.maxEntryValue)
	} else {
		endOffset := 64 - startBitOffset
		return int((int64(uint(b.data[startArrIndex])>>uint(startBitOffset)) | b.data[endArrIndex]<<endOffset) & b.maxEntryValue)
	}
}

func (b *bitArray) setBlock(index int64, value int) {
	if bits.Len(uint(value)) > b.BitsPerEntry() {
		*b = b.resize(b.bitsPerEntry+1, b.entrySize)
	}
	b.setAt(index, value)
}

func (b *bitArray) setAt(index int64, value int) {
	startOffset := index * int64(b.bitsPerEntry)
	startArrIndex := int(startOffset >> 6)
	endArrIndex := int(((index+1)*int64(b.bitsPerEntry) - 1) >> 6)
	startBitOffset := int(startOffset & 0x3F)
	b.data[startArrIndex] = b.data[startArrIndex] & ^(b.maxEntryValue<<startBitOffset) | (int64(value)&b.maxEntryValue)<<startBitOffset

	if startArrIndex != endArrIndex {
		endOffset := 64 - startBitOffset
		j1 := b.bitsPerEntry - endOffset
		b.data[endArrIndex] = int64(uint64(b.data[endArrIndex])>>uint64(j1))<<uint64(j1) | (int64(value)&b.maxEntryValue)>>endOffset

	}
}

func (b *bitArray) resize(bits, EntrySize int) bitArray {
	n := NewLitematicaBitArray(bits, EntrySize, nil)
	for i := 0; i < b.entrySize; i++ {
		n.setAt(int64(i), b.getAt(int64(i)))
	}
	return *n
}
