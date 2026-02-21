//@ostensibly-1/luxer

package luxer

import (
	"fmt"
	"log"
	"slices"
	"strconv"
	"strings"
	"time"
)

// globals
var RESERVED []string = []string{
	"and",
	"break",
	"do",
	"else",
	"elseif",
	"end",
	"false",
	"for",
	"function",
	"if",
	"in",
	"local",
	"nil",
	"not",
	"or",
	"repeat",
	"return",
	"then",
	"true",
	"until",
	"while",
}

var ESCAPE_CHAR_CONVERSION_TABLE map[string]string = map[string]string{
	"\\a":  "\a",
	"\\b":  "\b",
	"\\f":  "\f",
	"\\n":  "\n",
	"\\r":  "\r",
	"\\t":  "\t",
	"\\v":  "\v",
	"\\\\": "\\",
	"\\\"": "\"",
	"\\'":  "'",
	"\\[":  "\\[",
	"\\]":  "\\]",
}

var ALPHABET string = "QWERTYUIOPASDFGHJKLZXCVBNMqwertyuiopasdfghjklzxcvbnmo"
var DIGITS string = "1234567890"
var HEX_DIGITS string = "abcdef0123456789ABCDEF"
var IGNORE string = " \t\r\n"
var EASY_SYMBOLS string = "#%^*()+{}];:,/"
var ESCAPES string = "abfnrtv\\\"'[]"
var EQ_CHARS string = "=~<>"

// types
type TokenType int

const (
	TK_EOF TokenType = iota
	TK_NUMBER
	TK_STRING
	TK_NAME
	TK_SYMBOL
	TK_RESERVED
)

type Token struct {
	Type   TokenType
	Source string
}

func NewToken(t TokenType, s string) *Token {
	e := Token{
		Type:   t,
		Source: s,
	}

	return &e
}

type LexerObj struct {
	Tokens []Token
	Line   int
	Pos    int
	LPos   int
	Source string
}

func (self *LexerObj) Lex(src string) {
	self.Source = src

	//helpers
	AtEnd := func() bool {
		return self.Pos >= len(self.Source)
	}

	Peek := func(n int) string {
		return string(self.Source[self.Pos+n])
	}

	Adv := func() string {
		self.Pos = self.Pos + 1
		self.LPos = self.LPos + 1

		if Peek(-1) == "\n" {
			self.Line += 1
			self.LPos = 0
		}

		return Peek(-1)
	}

	AppendToken := func(ntok *Token) {
		self.Tokens = append(self.Tokens, *ntok)
	}

	TimeStart := time.Now()
	for !AtEnd() {
		Current := Peek(0)

		if strings.Contains(ALPHABET, Current) || Current == "_" {
			Name := Adv()

			// fahhhhhh
			for !AtEnd() && (strings.Contains(ALPHABET, Peek(0)) || strings.Contains(DIGITS, Peek(0)) || Peek(0) == "_") {
				Name = Name + Adv()
			}

			// fmt.Println(Current)

			if slices.Contains(RESERVED, Name) {
				AppendToken(NewToken(TK_RESERVED, Name))
			} else {
				AppendToken(NewToken(TK_NAME, Name))
			}
		} else if strings.Contains(IGNORE, Current) {
			Adv()
		} else if strings.Contains(DIGITS, Current) {
			Num := Adv()

			if Num == "0" && strings.ToLower(Peek(0)) == "x" {
				Num = Num + Adv()

				for !AtEnd() && strings.Contains(HEX_DIGITS, Peek(0)) {
					Num = Num + Adv()
				}
			} else {
				AllowDecimal := true
				LexNums := func() {
					for !AtEnd() && strings.Contains(DIGITS, Peek(0)) {
						Num = Num + Adv()
					}

					if strings.ToLower(Peek(0)) == "e" {
						Num = Num + Adv()
						AllowDecimal = false
					}

					if (Peek(0) == "+" || Peek(0) == "-") && AllowDecimal == false {
						Num = Num + Adv()
					}

					for !AtEnd() && strings.Contains(DIGITS, Peek(0)) {
						Num = Num + Adv()
					}
				}

				LexNums()

				if Peek(0) == "." && AllowDecimal != false {
					Num = Num + Adv()

					LexNums()
				}
			}

			AppendToken(NewToken(TK_NUMBER, Num))
		} else if strings.Contains(EASY_SYMBOLS, Current) {
			AppendToken(NewToken(TK_SYMBOL, Adv()))
		} else if Current == "\"" || Current == "'" {
			StrStart := Adv()
			var StrData strings.Builder

			for !AtEnd() && Peek(0) != StrStart {
				CurrentChar := Adv()

				if CurrentChar == "\\" {
					if AtEnd() {
						log.Fatal("unfinished string escape")
					}

					Next := Peek(0)

					if strings.Contains(DIGITS, Next) {
						NumStr := Adv()

						for i := 0; i < 2 && !AtEnd() && strings.Contains(DIGITS, Peek(0)); i++ {
							NumStr += Adv()
						}

						Val, _ := strconv.Atoi(NumStr)

						if Val > 255 {
							log.Fatal("escape char exceeded limit of 255")
						}

						StrData.WriteString(fmt.Sprintf("\\%d", Val))
					} else if strings.Contains(ESCAPES, Next) {
						EscChar := Adv()
						Converted := ESCAPE_CHAR_CONVERSION_TABLE["\\"+EscChar]

						StrData.WriteString(Converted)
					} else {
						log.Fatal("unsupported escape")
					}
				} else {
					StrData.WriteString(CurrentChar)
				}
			}

			if AtEnd() {
				log.Fatal("failed to lex string, no end")
			}

			Adv()

			AppendToken(NewToken(TK_STRING, StrData.String()))
		} else if Current == "[" {
			Start := Adv()

			if !AtEnd() && Peek(0) == "=" {
				Adv()
				EqCount := 1

				for !AtEnd() && Peek(0) == "=" {
					EqCount += 1
					Adv()
				}

				if !AtEnd() && Peek(0) != "[" {
					log.Fatal("failed init long str")
				}

				Adv()

				StrData := strings.Builder{}

				for !AtEnd() {
					Dump := strings.Builder{}
					if Peek(0) == "]" {
						// try to end

						Dump.WriteString(Adv())

						for i := 0; i < EqCount && !AtEnd() && Peek(0) == "="; i++ {
							Dump.WriteString(Adv())
						}

						if !AtEnd() && Peek(0) != "]" {
							StrData.WriteString(Dump.String())
							continue
						}

						Dump.WriteString(Adv())

						break
					}
					StrData.WriteString(Adv())
				}

				AppendToken(NewToken(TK_STRING, StrData.String()))
			} else if !AtEnd() && Peek(0) == "[" {
				Adv()

				StrData := strings.Builder{}

				for !AtEnd() {
					Dump := strings.Builder{}
					if Peek(0) == "]" {
						// try to end

						Dump.WriteString(Adv())

						if !AtEnd() && Peek(0) != "]" {
							StrData.WriteString(Dump.String())
							continue
						}

						Dump.WriteString(Adv())

						break
					}
					StrData.WriteString(Adv())
				}

				AppendToken(NewToken(TK_STRING, StrData.String()))
			} else {
				AppendToken(NewToken(TK_SYMBOL, Start))
			}
		} else if Current == "-" {
			Start := Adv()

			// i hate this section sm, my bad for this, first time writing in golang
			if !AtEnd() && Peek(0) == "-" {
				Adv()

				if !AtEnd() && Peek(0) == "[" {
					Adv()

					if !AtEnd() && Peek(0) == "=" {
						Adv()
						EqCount := 1

						for !AtEnd() && Peek(0) == "=" {
							EqCount += 1
							Adv()
						}

						if !AtEnd() && Peek(0) != "[" {
							log.Fatal("failed init long str")
						}

						Adv()

						for !AtEnd() {
							if Peek(0) == "]" {
								// try to end

								Adv()

								for i := 0; i < EqCount && !AtEnd() && Peek(0) == "="; i++ {
									Adv()
								}

								if !AtEnd() && Peek(0) != "]" {
									Adv()
									continue
								}

								Adv()

								break
							}
							Adv()
						}
					} else if !AtEnd() && Peek(0) == "[" {
						Adv()

						for !AtEnd() {
							if Peek(0) == "]" {
								// try to end

								Adv()

								if !AtEnd() && Peek(0) != "]" {
									Adv()
									continue
								}

								Adv()

								break
							}
							Adv()
						}
					} else {
						for !AtEnd() && Peek(0) != "\n" {
							Adv()
						}
					}
				} else {
					for !AtEnd() && Peek(0) != "\n" {
						Adv()
					}
				}
			} else {
				AppendToken(NewToken(TK_SYMBOL, Start))
			}
		} else if Current == "." {
			Start := Adv()

			if !AtEnd() && Peek(0) == "." {
				Start += Adv()
				if !AtEnd() && Peek(0) == "." {
					Start += Adv()
				}
			} else if !AtEnd() && strings.Contains(DIGITS, Peek(0)) {
				Num := Start

				for !AtEnd() && strings.Contains(DIGITS, Peek(0)) {
					Num = Num + Adv()
				}

				if strings.ToLower(Peek(0)) == "e" {
					Num = Num + Adv()
				}

				if Peek(0) == "+" || Peek(0) == "-" {
					Num = Num + Adv()
				}

				for !AtEnd() && strings.Contains(DIGITS, Peek(0)) {
					Num = Num + Adv()
				}

				AppendToken(NewToken(TK_NUMBER, Num))
				continue
			}

			AppendToken(NewToken(TK_SYMBOL, Start))
		} else if strings.Contains(EQ_CHARS, Current) {
			Src := Adv()

			if Src == "~" && Peek(0) != "=" {
				log.Fatal("fuck bitwise not")
			}

			if Peek(0) == "=" {
				Src += Adv()
			}

			AppendToken(NewToken(TK_SYMBOL, Src))
		} else {
			log.Fatal("unrecognized symbol, git gud")
		}
	}

	AppendToken(NewToken(TK_EOF, ""))

	log.Println("Luxer by Ostensibly-1")
	log.Printf("LU-xed source in %.2f!\n", time.Since(TimeStart).Seconds())
}

func NewLexer() *LexerObj {
	e := LexerObj{
		Line: 1,
		Pos:  0,
		LPos: 0,
	}

	return &e
}
