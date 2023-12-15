package main

import (
	"bytes"
	"fmt"
	"log"
	"os"
	"path/filepath"

	"github.com/chris-pikul/vole/parser"
)

type Lines [][]byte

func main() {
	log.Println("Vole")

	if len(os.Args) < 2 {
		log.Fatal("incorrect usage, expected: vole [file]")
	}

	path, err := filepath.Abs(os.Args[1])
	if err != nil {
		log.Fatal(err)
	}

	if filepath.Ext(path) == "" {
		path = path + ".vole"
	}

	file, err := os.ReadFile(path)
	if err != nil {
		log.Fatal(err)
	}

	ind := detectIndentation(bytes.Split(file, []byte{'\n'}))
	log.Printf("detected indentation of %s", ind)

	lexer := parser.NewLexer()
	lexer.Tokenize(file)
	lexer.DebugPrint()
}

type Indentation struct {
	Tabs bool
	Size uint8
}

func (ind Indentation) String() string {
	if ind.Tabs {
		return fmt.Sprintf("%d Tabs", ind.Size)
	}
	return fmt.Sprintf("%d Spaces", ind.Size)
}

func detectIndentation(data Lines) Indentation {
	data = stripDeadWeight(data)

	totalCnt := 0
	tabs := 0
	spaces := 0
	spaceLens := make(map[int]struct{})

	for _, ln := range data {
		if len(ln) > 0 {
			if ln[0] == '\t' {
				totalCnt++
				tabs++
			} else if ln[0] == ' ' {
				totalCnt++
				spaces++

				if num := indexByteNot(ln, ' '); num > 0 {
					spaceLens[num] = struct{}{}
				}
			}
		}
	}

	if tabs > spaces {
		return Indentation{Tabs: true, Size: 1}
	} else {
		spaceArr := make([]int, len(spaceLens))
		maxVal := 0
		{
			i := 0
			for k, _ := range spaceLens {
				spaceArr[i] = k
				if k > maxVal {
					maxVal = k
				}
			}
		}

		commons := make(map[int]int)
		for _, v := range spaceArr {
			cmn := lcd(v, maxVal)
			if cmn > 1 {
				commons[cmn]++
			}
		}

		common := 1
		most := 0
		for k, v := range commons {
			if v > most {
				most = v
				common = k
			}
		}

		return Indentation{Tabs: false, Size: uint8(common)}
	}
}

func stripDeadWeight(data Lines) Lines {
	inComment := false

	lines := make(Lines, 0)

	var tmp []byte
	for _, ln := range data {
		tmp = bytes.TrimSpace(ln)
		if len(tmp) >= 2 {
			if tmp[0] == '/' && tmp[1] == '/' {
				continue
			} else if tmp[0] == '/' && tmp[1] == '*' {
				inComment = true
				continue
			} else if inComment {
				if tmp[len(tmp)-2] == '*' && tmp[len(tmp)-1] == '/' {
					inComment = false
				}
			} else {
				lines = append(lines, ln)
			}
		} else if len(tmp) > 0 && !inComment {
			inline := bytes.Index(ln, []byte("//"))
			if inline != -1 {
				lines = append(lines, ln[0:inline])
			} else {
				lines = append(lines, ln)
			}
		}
	}

	return lines
}

func indexByteNot(src []byte, byt byte) int {
	for i, b := range src {
		if b != byt {
			return i
		}
	}
	return -1
}

func gcd(a, b int) int {
	if b == 0 {
		return a
	}
	return gcd(b, a%b)
}

func lcd(a, b int) int {
	if a > b {
		return (a / gcd(a, b)) * b
	}
	return (b / gcd(a, b)) * a
}
