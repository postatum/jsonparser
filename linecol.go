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
	for index, rune := range sj {
		fmt.Printf("at index '%v', rune is '%v', aka '%#v'\n", index, rune, string(rune))
		if rune == '\n' {
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
// Lines are numbered from 0, so offset 0 is at line 0, col 0.
//
func (li *LineIndex) OffsetToLineCol(offset int) (line int, bytecol int, runecol int) {
	fmt.Printf("\n top of OffsetToLineCol, offset=%v\n", offset)

	li.DebugDump()

	if offset >= len(li.JsonBlob) || offset < 0 {
		return -1, -1, -1
	}
	if offset == 0 {
		return 0, 0, 0
	}
	n := len(li.NewlinePos)

	if n == 0 {
		// no newlines in the indexed li.JsonBlob
		fmt.Printf("\n  no newlines in li.JsonBlob \n")
		return 0, offset, li.bytePosToRunePos(0, offset)
	}
	if offset >= li.NewlinePos[n-1] {
		// on the last line
		fmt.Printf("\n  on last line of li.JsonBlob, n=%v, li.NewlinePos[n-1==%v]==%v\n",
			n, n-1, li.NewlinePos[n-1])
		return n, offset - (li.NewlinePos[n-1] + 1), li.bytePosToRunePos(n, offset)
	}

	// binary search to locate the line using the li.NewlinePos index:
	//
	// sort.Search returns the smallest index i in [0, n) at which f(i) is true,
	// assuming that on the range [0, n), f(i) == true implies f(i+1) == true.
	//
	srch := sort.Search(n, func(i int) bool {
		r := (offset < li.NewlinePos[i])
		fmt.Printf("\n sort.Search on i = %v, with offset=%v, li.NewlinePos[i=%v] = %v, returning r=%v\n",
			i, offset, i, li.NewlinePos[i], r)
		return r
	})
	fmt.Printf("\n   srch=%v, n=%v, offset=%v\n", srch, n, offset)
	linestart := li.NewlinePos[srch-1] + 1
	fmt.Printf("\n linestart = %v, offset=%v\n", linestart, offset)
	return srch, offset - linestart, li.bytePosToRunePos(srch, offset)
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
	fmt.Printf("\n debug top of bytePosToRunePos, linenoz=%v, offset=%v\n", linenoz, offset)
	var beg int
	if linenoz > 0 {
		beg = li.NewlinePos[linenoz-1] + 1
	}
	fmt.Printf("\n debug bytePosToRunePos, linenoz=%v, offset=%v, beg=%v\n", linenoz, offset, beg)
	s := string(li.JsonBlob[beg : offset+1])
	fmt.Printf("\n debug bytePosToRunePos, s = '%s'\n", s)
	return utf8.RuneCountInString(s) - 1
}

func (li *LineIndex) DebugDump() {
	fmt.Println()
	for i := range li.NewlinePos {
		fmt.Printf("li.NewlinePos[i=%v]: %v\n", i, li.NewlinePos[i])
	}
	fmt.Println()
}
