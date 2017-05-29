package jsonparser

import (
	"fmt"
	"sort"
	"unicode/utf8"
)

// NewlineIndex holds the positions of all newlines
// in a given JSON blob. The JsonBlob must be utf8 text.
type LineIndex struct {
	JsonBlob   []byte
	NewlinePos []int
}

// NewLineIndex returns a new LineIndex whose
// NewlinePos member contains the byte-based
// locations of all newlines in the utf8 json.
func NewLineIndex(json []byte) *LineIndex {
	li := &LineIndex{
		JsonBlob:   json,
		NewlinePos: []int{},
	}
	li.FindNewlines()
	return li
}

// FindNewlines locates the newlines in the utf8 li.JsonBlob.
func (li *LineIndex) FindNewlines() {

	li.NewlinePos = []int{}

	// convert json to a string, in order to range over runes.
	// c.f. https://blog.golang.org/strings
	sj := string(li.JsonBlob)
	for index, rn := range sj {
		if rn == '\n' {
			li.NewlinePos = append(li.NewlinePos, index)
		}
	}
}

// OffsetToLineCol returns the line and column for a given offset,
// provided that li has been constructed by NewLineIndex so that
// li.NewlinePos is valid. It does so by binary search for offset
// on li.NewlinePos, so its time complexity is O(log q) where q
// is the number of newlines in li.JsonBlob.
//
// Note that bytecol is the byte index of the offset on the line,
// while runecol is the utf8 rune index on the line.
//
// OffsetToLineCol returns line of -1 if offset is out of bounds.
//
// Lines are numbered from 1, so offset 0 is at line 1, col 1.
//
func (li *LineIndex) OffsetToLineCol(offset int) (line int, bytecol int, runecol int) {
	// li.DebugDump()

	// Should this check be ">="? Is there a column after offset of full string length?
	if offset > len(li.JsonBlob) || offset < 0 {
		return -1, -1, -1
	}
	if offset == 0 {
		return incr(0, 0, 0)
	}
	numNewLines := len(li.NewlinePos)

	// No newlines in the indexed li.JsonBlob
	if numNewLines == 0 {
		return incr(0, offset, li.bytePosToRunePos(0, offset))
	}
	// On the last line
	if offset > li.NewlinePos[numNewLines-1] {
		return incr(numNewLines, offset - (li.NewlinePos[numNewLines-1] + 1), li.bytePosToRunePos(numNewLines, offset))
	}
	// Offset of whole string length = last column
	if offset == len(li.JsonBlob) {
		return incr(numNewLines, offset - (li.NewlinePos[numNewLines-1] + 1), li.bytePosToRunePos(numNewLines, offset))
	}

	// Binary search to locate the line using the li.NewlinePos index:
	//
	// sort.Search returns the smallest index i in [0, n) at which f(i) is true,
	// assuming that on the range [0, n), f(i) == true implies f(i+1) == true.
	//
	srch := sort.Search(numNewLines, func(i int) bool {
		r := (offset <= li.NewlinePos[i])
		return r
	})
	linestart := li.NewlinePos[srch-1] + 1
	return incr(srch, offset - linestart, li.bytePosToRunePos(srch, offset))
}

func incr(x, y, z int) (int, int, int) {
	return x + 1, y + 1, z + 1
}

// bytePosToRunePos expects linenoz to be zero-based line-number
// on which offset falls; i.e. that offset >= li.NewlinePos[linenoz-1];
// and offset < li.NewlinePos[linenoz] assuming linenoz is valid.
//
// It then returns the character (utf8 rune) position of the
// offset on that line.
//
// Since it must parse bytes into utf8 characters, the time complexity of
// bytePosToRunePos is O(length of the line).
//
func (li *LineIndex) bytePosToRunePos(linenoz int, offset int) int {
	var beg int
	if linenoz > 0 {
		beg = li.NewlinePos[linenoz-1] + 1
	}
	s := string(li.JsonBlob[beg : offset+1])
	return utf8.RuneCountInString(s) - 1
}

func (li *LineIndex) DebugDump() {
	fmt.Println()
	for i := range li.NewlinePos {
		fmt.Printf("li.NewlinePos[i=%v]: %v\n", i, li.NewlinePos[i])
	}
	fmt.Println()
}
