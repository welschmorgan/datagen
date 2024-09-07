package term

import (
	"bytes"
)

type Code string

type Style []Code

const (
	Escape Code = "\x1b["

	ShowCursor Code = "?25h"
	HideCursor Code = "?25l"

	Reset     Code = "0m"
	DefaultFg Code = "39m"
	DefaultBg Code = "49m"

	LightBlackFg   Code = "0;30m"
	LightRedFg     Code = "0;31m"
	LightGreenFg   Code = "0;32m"
	LightYellowFg  Code = "0;33m"
	LightBlueFg    Code = "0;34m"
	LightMagentaFg Code = "0;35m"
	LightCyanFg    Code = "0;36m"
	LightWhiteFg   Code = "0;37m"

	LightBlackBg   Code = "0;40m"
	LightRedBg     Code = "0;41m"
	LightGreenBg   Code = "0;42m"
	LightYellowBg  Code = "0;43m"
	LightBlueBg    Code = "0;44m"
	LightMagentaBg Code = "0;45m"
	LightCyanBg    Code = "0;46m"
	LightWhiteBg   Code = "0;47m"

	DarkBlackFg   Code = "1;30m"
	DarkRedFg     Code = "1;31m"
	DarkGreenFg   Code = "1;32m"
	DarkYellowFg  Code = "1;33m"
	DarkBlueFg    Code = "1;34m"
	DarkMagentaFg Code = "1;35m"
	DarkCyanFg    Code = "1;36m"
	DarkWhiteFg   Code = "1;37m"

	DarkBlackBg   Code = "1;40m"
	DarkRedBg     Code = "1;41m"
	DarkGreenBg   Code = "1;42m"
	DarkYellowBg  Code = "1;43m"
	DarkBlueBg    Code = "1;44m"
	DarkMagentaBg Code = "1;45m"
	DarkCyanBg    Code = "1;46m"
	DarkWhiteBg   Code = "1;47m"
)

func Styled(style ...Code) Style {
	return Style(style)
}

func (s Style) String() string {
	buf := bytes.NewBuffer([]byte{})
	buf.Write([]byte(Escape))
	for _, code := range s {
		buf.Write([]byte(code))
	}
	return buf.String()
}
