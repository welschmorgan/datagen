package term

import (
	"bytes"
	"fmt"
	"strings"
)

type ProgressBar struct {
	size    int
	value   float32
	barChar rune
}

const DefaultBarChar rune = '▉'
const DefaultSize int = 10

func DefaultProgressBar() *ProgressBar {
	return NewProgressBar(DefaultSize, DefaultBarChar)
}

func NewProgressBar(size int, barChar rune) *ProgressBar {
	return &ProgressBar{size: size, barChar: barChar, value: 0.0}
}

func (b *ProgressBar) SetValue(value float32) {
	b.value = value
}

func (b *ProgressBar) Value() float32 {
	return b.value
}

func (b *ProgressBar) SetBarChar(value rune) {
	b.barChar = value
}

func (b *ProgressBar) BarChar() rune {
	return b.barChar
}

func (b *ProgressBar) String() string {
	numBars := int(b.value * float32(b.size))
	restBars := b.size - numBars
	buf := bytes.NewBuffer([]byte{})
	buf.WriteString(strings.Repeat(fmt.Sprintf("%c", b.barChar), numBars))
	buf.WriteString(strings.Repeat(" ", restBars))
	return buf.String()
}
