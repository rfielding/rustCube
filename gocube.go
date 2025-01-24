package main

import (
	"fmt"
	"strings"
)

/*
  This is a Go implementation just so that I can get it done.
  It is very unproductive to use Rust for writing a parser.
  I will work on the Rust version to learn Rust, but the Go version
  to get the parsing to work right.

  In Rust, I am having to figure out a state machine to parse with,
  just to avoid all of the copy/move stuff that seems so painfully
  unnecessary to just get a working language parser.
*/

type Cube struct {
	FaceCount  int
	FacePeriod int
	Adj        map[string][]string
	Faces      []string
	Opposite   map[string]string
	Stickers   map[string]string // urf -> u, ur = u, rfu -> r, ...
}

func NewCube() *Cube {
	cube := &Cube{
		FaceCount:  6,
		FacePeriod: 4,
		// orderings of faces
		Faces: []string{"u", "r", "f", "d", "l", "b"},
		// opposite faces calculated
		Opposite: map[string]string{
			"u": "d",
			"r": "l",
			"f": "b",
			"d": "u",
			"l": "r",
			"b": "f",
		},
		// adjacencies are counter-clockwise, so that swaps produce a clockwise turn
		Adj: map[string][]string{
			"u": {"f", "r", "b", "l"},
			"r": {"u", "f", "d", "b"},
			"f": {"u", "l", "d", "r"},
			"d": {"f", "l", "b", "r"},
			"l": {"u", "b", "d", "f"},
			"b": {"u", "r", "d", "l"},
		},
		// state
		Stickers: make(map[string]string),
	}
	// i,j,k are strings to located faces
	// fi finds turn face, fj is an adjacent face to find j and k
	// corners must be counter-clockwise, or everything fails
	//
	//   | i | k
	//     j
	//
	for fi := 0; fi < cube.FaceCount; fi++ {
		i := cube.Faces[fi]
		cube.Stickers[i] = i
		for fj := 0; fj < cube.FacePeriod; fj++ {
			j := cube.Adj[i][fj]
			k := cube.Adj[i][(fj+1)%4]
			// corner i orbit
			cube.Stickers[i+k+j] = i
			// edge i orbit
			cube.Stickers[i+j] = i
		}
	}
	// make sure that it satisfied solved cube invariants
	for s, v := range cube.Stickers {
		if v[0] != s[0] {
			panic(fmt.Sprintf("stickers should start with face name: %s vs %s", s, v))
		}
	}
	if cube.Stickers["bul"] != "" {
		if cube.Stickers["blu"] == "b" {
			panic("stickers should be clockwise, so bul should be blu")
		}
	}
	return cube
}

// stdio side-effect
func (cube *Cube) Draw(cmd string, repeats int) {
	s := func(s string) string {
		v := cube.Stickers[s]
		if v == "" {
			panic(fmt.Sprintf("sticker is not mapped: %s\n%v", s, cube.Stickers))
		}
		return v
	}

	// draw a 3x3x3 cube now, with clockwise corners
	fmt.Printf("      %s%s%s      \n",
		s("bul"), s("bu"), s("bru"),
	)
	fmt.Println()
	fmt.Printf("      %s%s%s      \n",
		s("ulb"), s("ub"), s("ubr"),
	)
	fmt.Printf("      %s%s%s      \n",
		s("ul"), s("u"), s("ur"),
	)
	fmt.Printf("      %s%s%s      \n",
		s("ufl"), s("uf"), s("urf"),
	)
	fmt.Println()
	fmt.Printf("%s %s%s%s %s%s%s %s%s%s %s\n",
		s("bul"),
		s("lbu"), s("lu"), s("luf"),
		s("flu"), s("fu"), s("fur"),
		s("rfu"), s("ru"), s("rub"),
		s("bru"),
	)
	fmt.Printf("%s %s%s%s %s%s%s %s%s%s %s\n",
		s("bl"),
		s("lb"), s("l"), s("lf"),
		s("fl"), s("f"), s("fr"),
		s("rf"), s("r"), s("rb"),
		s("br"),
	)
	fmt.Printf("%s %s%s%s %s%s%s %s%s%s %s\n",
		s("bld"),
		s("ldb"), s("ld"), s("lfd"),
		s("fdl"), s("fd"), s("frd"),
		s("rdf"), s("rd"), s("rbd"),
		s("bdr"),
	)
	fmt.Println()
	fmt.Printf("      %s%s%s      \n",
		s("dlf"), s("df"), s("dfr"),
	)
	fmt.Printf("      %s%s%s      \n",
		s("dl"), s("d"), s("dr"),
	)
	fmt.Printf("      %s%s%s      \n",
		s("dbl"), s("db"), s("drb"),
	)
	fmt.Println()
	fmt.Printf("      %s%s%s      \n",
		s("bld"), s("bd"), s("bdr"),
	)
	fmt.Println()
	fmt.Printf("cmd: %s x %d\n", cmd, repeats)

}

// 1 turn of  and maybe center at face i,
//
//	physical parts: ru~ur, rub~ubr~bru
func (cube *Cube) Turn1(f string, center bool) {
	s := func(s string) string {
		v := cube.Stickers[s]
		if v == "" {
			panic(fmt.Sprintf("sticker is not mapped: %s\n%v", s, cube.Stickers))
		}
		return v
	}

	swap := func(a string, b string) {
		// use s(a) to verify that value is not null
		tmp := s(a)
		cube.Stickers[a] = s(b)
		cube.Stickers[b] = tmp
		//fmt.Printf("%s %s\n", a, b)
	}

	// faces have a period of 4, move their stickers
	//
	//  k | f | j
	//    -----
	//      i
	//
	// edge:   fi,if -> fj,jf
	// corner: fik,ikf,kfi -> fji,jif,fji
	//
	// note that because we swap in pairs, it's one-less than period
	//
	for fi := 0; fi < cube.FacePeriod-1; fi++ {
		k := cube.Adj[f][(fi+3)%cube.FacePeriod] //behind fi
		i := cube.Adj[f][(fi+0)%cube.FacePeriod] //at fi
		j := cube.Adj[f][(fi+1)%cube.FacePeriod] //ahead fi

		e0a := f + i
		e1a := i + f
		e0b := f + j
		e1b := j + f
		swap(e0a, e0b)
		swap(e1a, e1b)

		c0a := f + i + k
		c1a := i + k + f
		c2a := k + f + i
		c0b := f + j + i
		c1b := j + i + f
		c2b := i + f + j
		swap(c0a, c0b)
		swap(c1a, c1b)
		swap(c2a, c2b)

		if center {
			m0a := i
			m0b := j
			swap(m0a, m0b)
			e0a := i + k
			e1a := k + i
			e0b := j + i
			e1b := i + j
			swap(e0a, e0b)
			swap(e1a, e1b)
		}
	}
}

// turn a face *count* times, all cube or just a face
func (cube *Cube) Turn(i string, count int) {
	//
	all := cube.shouldTurnWholeCube(i)
	i = strings.ToLower(i)

	for count < 0 {
		count += cube.FacePeriod
	}
	count = count % cube.FacePeriod

	// turn a face count times.
	if all {
		// turn face and center
		for n := 0; n < count; n++ {
			cube.Turn1(i, true)
		}
		// triple is negative, needed to turn back face
		ncount := ((cube.FacePeriod - 1) * count) % cube.FacePeriod
		for n := 0; n < ncount; n++ {
			cube.Turn1(cube.Opposite[i], false)
		}
	} else {
		for n := 0; n < count; n++ {
			cube.Turn1(i, false)
		}
	}
}

func (cube *Cube) shouldTurnWholeCube(f string) bool {
	if f == "U" || f == "R" || f == "F" || f == "D" || f == "L" || f == "B" {
		return true
	}
	return false
}

func (cube *Cube) shouldTurnCube(f string) bool {
	if f == "u" || f == "r" || f == "f" || f == "d" || f == "l" || f == "b" {
		return true
	}
	return false
}

func (cube *Cube) help() {
	fmt.Printf("turn a face: u r f d l b\n")
	fmt.Printf("turn cube:   U R F D L B\n")
	fmt.Printf("reverse turn '/', like: /u\n")
	fmt.Printf("example: u r /u /r\n")
	fmt.Printf("example: UUUU returns to where it started\n")
	fmt.Printf("help: ?\n")
	fmt.Printf("new cube: n\n")
}

type Node struct {
	Face       string
	Negate     bool
	Commutator bool
	Arr        []Node
	Repeat     int
}

func (node Node) Print() string {
	v := ""
	if node.Negate {
		v += "/"
	}
	if node.Arr != nil {
		if node.Commutator {
			v += "["
		} else {
			v += "("
		}
		for i, n := range node.Arr {
			if i > 0 {
				v += " "
			}
			v += n.Print()
		}
		if node.Commutator {
			v += "]"
		} else {
			v += ")"
		}
	} else {
		v += node.Face
	}
	if node.Repeat != 0 {
		v += fmt.Sprintf("%d", node.Repeat)
	}
	return v
}

// parseParentheses parses the input string and constructs a nested Node structure.
func Parse(input string) Node {
	stack := [][]Node{{}}
	wasNegated := false
	negParens := false
	for i := 0; i < len(input); i++ {
		char := input[i]

		switch char {
		case '(', '[':
			stack = append(
				stack,
				[]Node{},
			)
		case ')', ']':
			if len(stack) > 1 {
				last := stack[len(stack)-1]
				stack = stack[:len(stack)-1]
				stack[len(stack)-1] = append(
					stack[len(stack)-1],
					Node{
						Commutator: char == ']',
						Arr:        last,
						Negate:     negParens,
					},
				)
			}
			negParens = false
		case '/':
			// use it to set negate on next token
			wasNegated = true
			negParens = true
			continue
		case 'U', 'R', 'F', 'D', 'L', 'B', 'u', 'r', 'f', 'd', 'l', 'b':
			face := char
			top := len(stack) - 1
			stack[top] = append(
				stack[top],
				Node{
					Face:   string(face),
					Negate: wasNegated,
				},
			)
		case '0', '1', '2', '3', '4', '5', '6', '7', '8', '9':
			// look ahead to complete the number, and look back to write the repeat
			num := 0
			numStop := i
			for numStop < len(input) && '0' < input[numStop] && input[numStop] <= '9' {
				num = 10*num + int(input[numStop]-'0')
				numStop++
			}
			i = numStop - 1
			if 0 < i {
				top := len(stack) - 1
				if len(stack[top]) > 0 {
					stack[top][len(stack[top])-1].Repeat = num
				}
			}
		}
		wasNegated = false
	}
	return Node{Arr: stack[0]}
}

func reverse[T any](s []T) {
	for i, j := 0, len(s)-1; i < j; i, j = i+1, j-1 {
		s[i], s[j] = s[j], s[i]
	}
}

func (cube *Cube) Execute(node Node, negates int) {
	repeat := 1
	if node.Repeat != 0 {
		repeat = node.Repeat
	}
	if node.Negate {
		negates++
	}
	if node.Arr != nil {
		for i := 0; i < repeat; i++ {
			theList := node.Arr
			if node.Negate && !node.Commutator {
				reverse(theList)
			}
			for _, cmd := range node.Arr {
				cube.Execute(cmd, negates)
			}
			if node.Commutator {
				for _, cmd := range node.Arr {
					cube.Execute(cmd, negates+1)
				}
			}
		}
	} else {
		if node.Face != "" {
			turn := repeat
			turn = turn * (1 - 2*(negates%2))
			fmt.Printf("do: %s %d\n", node.Face, turn)
			cube.Turn(node.Face, turn)
		}
	}
}

func (cube *Cube) Loop() {
	// loop to get and anlyze a line and draw the screen
	cmd := ""
	repeats := 0
	prevCmd := ""
	cube.help()
	for {
		cube.Draw(cmd, repeats)
		fmt.Scanln(&cmd)

		if cmd == "q" {
			break
		}

		if cmd == "n" {
			cube = NewCube()
			continue
		}

		if cmd == "?" {
			cube.help()
			continue
		}

		if cmd == prevCmd || cmd == "" {
			if cmd == "" {
				cmd = prevCmd
			}
			repeats = repeats + 1
		} else {
			repeats = 1
		}
		prevCmd = cmd

		nodes := Parse(cmd)

		fmt.Printf("%s\n", nodes.Print())

		cube.Execute(nodes, 0)
	}
}

func main() {
	cube := NewCube()
	cube.Loop()
}
