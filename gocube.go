package main

import (
	"bufio"
	"flag"
	"fmt"
	"os"
	"strings"
)

var PostTest = flag.Bool("postTest", false, "post test on start")

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
	Stickers   map[string]string
	History    []map[string]string
}

type Node struct {
	Face       string
	Negate     bool
	Commutator bool
	Conjugated bool
	Arr        []Node
	Repeat     int
	Reflection string
}

var UseAnsi = true

// tests happen in an order that finds most primitive bugs that should cause
// later cases to fail, and teaches user how to think about the algebra
// when looking at examples
var EqTest = [][]string{
	{"uuuu -- face turn period 4", "u4", "u2 u2", "u u3", ""},
	{"UUUU -- cube turn period 4", "U4", "U2 U2", "U U3", ""},
	{"(fr /f/r)6 -- commutator period 6 is important", "[f r]6", ""},
	{"(fr /f/r)3 (f r /f /r)3", ""},
	{"[fr]2 [fr]4 -- all adjacent face commuators have period 6", ""},
	{"[fr]3 [fr]3", ""},
	{"(fr)/(rf)   -- a raw commutator"},
	{"{fr}        -- a conjugate, wrap r in f.", "f r /f"},
	{"{fr}/{fr}   -- conjugate identity", ""},
	{"{f [ru]}    -- orient colors up in u face after bottom 2 layers done", "f [ru] /f"},
	{"((fr)/(fr))6 -- period 6. adjacent face commutators are important!", ""},
	{"/(u /(r /f))", "/(u f /r)", "r /f /u"},
	{"/[fd]", "[df]"},
	{"[fr]/[fr]", ""},
	{"[fr][rf]", "[fr]/[fr]", ""},
	{"[/r d] d2 [f/d]  -- after solved u layer, middle edge insert"},
	{"RR               -- after one side solved, flip cube upside down yellow center is u face"},
	{"f [ur] /f        -- get all u edge colors into u"},
	{"r u /r u r u2 /r -- swap edge pairs while leaving u face in u"},
	{"[[fr]3 u]        -- last layer edge cycle", "[((fr)/(rf))3 u]"},
	{"[[fd]2 u]        -- last layer edge twist", "[fd]2 u /[fd]2 /u"},
	{"{/r d} [d f]     -- edge after u solved ", "/r d r d f /d /f"},
	{"{l /d} [/d /f]   -- edge after u is solved", "l /d /l /d /f d f"},
	{"{f {ru}}         -- turn u inside [fr] center"},
	{"(x {l /d} [/d /f] )   -- mirror image a move accros axis R. negate all faces and swap f-b,l-r,u~d. to reuse moves.", "{/rd}[df]"},
}

func strippedComment(s string) string {
	return strings.Trim(strings.Split(s, "-")[0], " ")
}

func sameMeaning(s string) string {
	// strip spaces after stripped comments
	s = strippedComment(s)
	// strip spaces
	s = strings.Replace(s, " ", "", -1)
	return s
}

func (cube *Cube) assert(s string) {
	if *PostTest {
		panic(s)
	} else {
		cube.PrintRed(s)
	}
}

/*
It seems a little strange to do this instead of standard Go test, but
I will include integration tests if I am to provide internal parameters,
for example to hook up to OpenAI and ask it to solve cubes.

But any parameters not compiled in would be cause to do a PostTest.
*/
func (cube *Cube) PostTest() {
	fmt.Printf("running post test\n")
	checkInterpretation := func(s string, theCube *Cube) Node {
		fmt.Printf("checkInterpretation: %s\n", s)
		expect := "(" + s + ")"
		parsed, err := theCube.Parse(s)
		if err != nil {
			cube.assert(fmt.Sprintf("parse error on example %s: %s\n", s, err))
		}
		got := parsed.Print()
		got = sameMeaning(got)
		expect = sameMeaning(expect)
		if expect != got {
			cube.assert(fmt.Sprintf("expect interpretation of %s to be: %s\n", expect, got))
		}
		return parsed
	}

	checkExecution := func(parsed Node, theCube *Cube) {
		fmt.Printf("checkExecution: %s\n", parsed.Print())
		execution, err := theCube.ExecuteCommand(parsed)
		if err != nil {
			cube.assert(fmt.Sprintf("execute error on example %s: %s\n", parsed.Print(), err))
		}
		_ = execution
	}

	checkInvertability := func(s string) {
		fmt.Printf("checkInvertability: %s\n", s)
		if len(s) < 3 {
			return
		}
		sNot := s
		if s[0] == '/' && (s[1] == '(' || s[1] == '[') {
			sNot = s[1:]
		} else {
			sNot = "/(" + s + ")"
		}
		c1 := NewCube()
		node, err := c1.Parse(s)
		if err != nil {
			cube.assert(fmt.Sprintf("parse error on example invertability chech %s: %s\n", s, err))
		}
		ex1, err := c1.ExecuteCommand(node)
		if err != nil {
			cube.assert(fmt.Sprintf("execute error on example invertability chech %s: %s\n", s, err))
		}
		node, err = c1.Parse(sNot)
		if err != nil {
			cube.assert(fmt.Sprintf("parse error on example invertability chech %s: %s\n", sNot, err))
		}
		ex2, err := c1.ExecuteCommand(node)
		if err != nil {
			cube.assert(fmt.Sprintf("execute error on example invertability chech %s: %s\n", sNot, err))
		}
		for k, v := range c1.Stickers {
			if string(k[0]) != v {
				cube.assert(
					fmt.Sprintf(
						"inverse check: %s not inverted by %s.\nfwd: %s\nrev: %s\n",
						s,
						sNot,
						ex1,
						ex2,
					),
				)
			}
		}
	}

	for i := range EqTest {
		// check the INTERPRETATION after a parse
		s := strippedComment(EqTest[i][0])
		checkInvertability(s)

		cube1 := NewCube()
		parsed := checkInterpretation(s, cube1)
		checkExecution(parsed, cube1)

		// compare next string cubes to current cube state.
		// stickers should be the same to pass the test.
		for j := 1; j < len(EqTest[i]); j++ {
			s2 := strippedComment(EqTest[i][j])
			checkInvertability(s2)

			cube2 := NewCube()
			parsed2 := checkInterpretation(s2, cube2)
			checkExecution(parsed2, cube2)

			// compare stickers to make sure they are equivalent as a parse
			to := s2
			if to == "" {
				to = "()"
			}
			fmt.Printf("checkEquality: %s == %s\n", s, to)
			for k := range cube1.Stickers {
				got := cube1.Stickers[k]
				expected := cube2.Stickers[k]
				if got != expected {
					cube.assert(
						fmt.Sprintf(
							"stickers should be the same in %s: sticker %s got %s instead of %s\n",
							s,
							k,
							got,
							expected,
						),
					)
				}
			}
		}
	}
	fmt.Printf("post test complete\n\n")
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
		// state of solve
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

// Draw with an allowance for 2 cubes side-by-side
func Draw(cmd string, repeats int, cubes []*Cube) {
	fullMask := "%s  %s%s%s  %s%s%s  %s%s%s  %s"
	edgeMask := "            %s%s%s            "

	s := func(cube *Cube, x string) string {
		v := cube.Stickers[x]
		if v == "" {
			panic(fmt.Sprintf("sticker is not mapped: %s\n%v", x, cube.Stickers))
		}
		// fg colors: 30 black, 31 red, 32 green, 33 yellow, 34 blue, 35 magenta, 36 cyan, 37 white
		// bg colors: 40 black, 41 red, 42 green, 43 yellow, 44 blue, 45 magenta, 46 cyan, 47 white
		if UseAnsi {
			switch v {
			case "u":
				v = fmt.Sprintf("\u001b[1;47;30m  \u001b[0m")
			case "r":
				v = fmt.Sprintf("\u001b[1;44;30m  \u001b[0m")
			case "f":
				v = fmt.Sprintf("\u001b[1;41;30m  \u001b[0m")
			case "d":
				v = fmt.Sprintf("\u001b[1;43;30m  \u001b[0m")
			case "l":
				v = fmt.Sprintf("\u001b[1;42;30m  \u001b[0m")
			case "b":
				v = fmt.Sprintf("\u001b[1;45;37m  \u001b[0m")
			}
		} else {
			switch v {
			case "u":
				v = "u "
			case "r":
				v = "r "
			case "f":
				v = "f "
			case "d":
				v = "d "
			case "l":
				v = "l "
			case "b":
				v = "b "
			}
		}
		return v
	}

	draw := func(mask string, args ...string) {
		if mask == "" {
			fmt.Println()
			return
		}
		for c := 0; c < len(cubes); c++ {
			a := make([]any, 0)
			if mask == edgeMask {
				for i := 0; i < 3; i++ {
					a = append(a, s(cubes[c], args[i]))
				}
				fmt.Printf(edgeMask, a...)
				fmt.Printf("      ")
			}
			if mask == fullMask {
				for i := 0; i < 11; i++ {
					a = append(a, s(cubes[c], args[i]))
				}
				fmt.Printf(fullMask, a...)
				fmt.Printf("      ")
			}
		}
		fmt.Println()
	}

	draw(edgeMask,
		"bld", "bd", "bdr",
	)
	draw(edgeMask,
		"bl", "b", "br",
	)
	draw(edgeMask,
		"bul", "bu", "bru",
	)
	draw("")
	draw(edgeMask,
		"ulb", "ub", "ubr",
	)
	draw(edgeMask,
		"ul", "u", "ur",
	)
	draw(edgeMask,
		"ufl", "uf", "urf",
	)
	draw("")
	draw(fullMask,
		"bul",
		"lbu", "lu", "luf",
		"flu", "fu", "fur",
		"rfu", "ru", "rub",
		"bru",
	)
	draw(fullMask,
		"bl",
		"lb", "l", "lf",
		"fl", "f", "fr",
		"rf", "r", "rb",
		"br",
	)
	draw(fullMask,
		"bld",
		"ldb", "ld", "lfd",
		"fdl", "fd", "frd",
		"rdf", "rd", "rbd",
		"bdr",
	)
	draw("")
	draw(edgeMask,
		"dlf", "df", "dfr",
	)
	draw(edgeMask,
		"dl", "d", "dr",
	)
	draw(edgeMask,
		"dbl", "db", "drb",
	)
	draw("")
	draw(edgeMask,
		"bld", "bd", "bdr",
	)

	draw("")

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
	// make sure that there are still 9 stickers of every color!
	counts := make(map[byte]int)
	for _, v := range cube.Stickers {
		color := v[0]
		counts[color]++
	}
	for i := 0; i < cube.FaceCount; i++ {
		if counts[cube.Faces[i][0]] != 9 {
			panic(fmt.Sprintf("face %s has %d stickers", cube.Faces[i], counts[cube.Faces[i][0]]))
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

func (cube *Cube) facesString(upperCase bool) string {
	if !UseAnsi {
		return " U  R  F  D  L  B"
	}
	ansiColors := map[string]string{
		"u": "\u001b[1;37m%s\u001b[0m",
		"r": "\u001b[1;34m%s\u001b[0m",
		"f": "\u001b[1;31m%s\u001b[0m",
		"d": "\u001b[1;33m%s\u001b[0m",
		"l": "\u001b[1;32m%s\u001b[0m",
		"b": "\u001b[1;35m%s\u001b[0m",
	}

	uc := func(s string) string {
		if upperCase {
			return strings.ToUpper(s)
		}
		return s
	}

	// use the middle piece location to find the color
	// for u r f d l b
	return fmt.Sprintf(" %s  %s  %s  %s  %s  %s",
		fmt.Sprintf(ansiColors[cube.Stickers["u"]], uc("u")),
		fmt.Sprintf(ansiColors[cube.Stickers["r"]], uc("r")),
		fmt.Sprintf(ansiColors[cube.Stickers["f"]], uc("f")),
		fmt.Sprintf(ansiColors[cube.Stickers["d"]], uc("d")),
		fmt.Sprintf(ansiColors[cube.Stickers["l"]], uc("l")),
		fmt.Sprintf(ansiColors[cube.Stickers["b"]], uc("b")),
	)
}

func (cube *Cube) colorStr(color int, s string) string {
	if !UseAnsi {
		return s
	}
	return fmt.Sprintf("\u001b[1;%dm%s\u001b[0m", color, s)
}

func (cube *Cube) Help() {
	cube.PrintRed("-----BEGIN HELP-----\n")
	fmt.Printf("run inside rlwrap for better keyboard handling! ctrl-r to search commands.\n")
	fmt.Printf("from gocube.go file at: %s\n", cube.colorStr(32, "https://github.com/rfielding/rustCube"))
	fmt.Printf("conventions: Up Right Front Down Left Back\n")
	fmt.Printf("reverse a turn with '/', like: /u\n")
	fmt.Printf("commutators have period 6: ((ru)/(ur))6 => ()\n")
	fmt.Printf("negate: /(rf) = /f /r\n")
	fmt.Printf("commutator: [fr] => f r /f /r\n")
	fmt.Printf("conjugate: {ru} => r u /r\n")
	fmt.Printf("mirror across an axis: negate all faces and swap names on axis. ex: x[[fr]3u]=[[/f/l]3/u] \n")
	fmt.Println()
	for i := range EqTest {
		for j := 0; j < len(EqTest[i]); j++ {
			if j == 0 {
				fmt.Printf("%s ", cube.colorStr(34, "example:"))
			} else {
				fmt.Printf(" == ")
			}
			v := EqTest[i][j]
			if v == "" {
				v = "()"
			}
			fmt.Printf("%s", v)
		}
		fmt.Println()
	}
	fmt.Printf("%s nru         -- start from new cube, then ru\n", cube.colorStr(34, "example:"))
	fmt.Printf("%s n(fdrfdbl)5 -- for a deterministic scramble, you can find in rlwrap history\n", cube.colorStr(34, "example:"))
	fmt.Println()
	fmt.Printf("help: ? or h\n")
	fmt.Printf("new cube: n\n")
	fmt.Printf("toggle ansi colors: a\n")
	fmt.Printf("test: run tests on expressions\n")
	fmt.Printf("quit: q\n")
	fmt.Println()
	fmt.Printf("turn a face: %s\n", cube.facesString(false))
	fmt.Printf("turn cube:   %s\n", cube.facesString(true))
	fmt.Printf("pop move off history (undo): p\n")
	fmt.Printf("swap cubes: s\n")
	fmt.Printf("startup test flag: -postTest\n")
	cube.PrintRed("-----END HELP-----\n")
}

func (node Node) Print() string {
	v := ""
	if node.Negate {
		v += "/"
	}
	if node.Arr != nil {
		if node.Commutator {
			if node.Conjugated {
				v += "{"
			} else {
				v += "["
			}
		} else {
			v += "("
		}
		v += node.Reflection
		if len(node.Reflection) > 0 {
			v += " "
		}
		for i, n := range node.Arr {
			if i > 0 {
				v += " "
			}
			v += n.Print()
		}
		if node.Commutator {
			if node.Conjugated {
				v += "}"
			} else {
				v += "]"
			}
		} else {
			v += ")"
		}
	} else {
		v += node.Face
	}
	if node.Repeat != 0 {
		if node.Repeat != 1 {
			v += fmt.Sprintf("%d", node.Repeat)
		}
	}
	return v
}

// parseParentheses parses the input string and constructs a nested Node structure.
func (cube *Cube) Parse(input string) (Node, error) {
	// string comments with --
	input = sameMeaning(input)

	// parenthesis balance
	openParenCount := 0
	closeParenCount := 0
	parenBalance := 0
	openBracketCount := 0
	closeBracketCount := 0
	bracketBalance := 0
	openConjugateCount := 0
	closeConjugateCount := 0
	conjugateBalance := 0
	for i := 0; i < len(input); i++ {
		char := input[i]
		switch char {
		case '(':
			openParenCount++
			parenBalance++
		case ')':
			closeParenCount++
			parenBalance--
			if parenBalance < 0 {
				return Node{}, fmt.Errorf("unbalanced parentheses, (, and )")
			}
		case '[':
			openBracketCount++
			bracketBalance++
		case ']':
			closeBracketCount++
			bracketBalance--
			if bracketBalance < 0 {
				return Node{}, fmt.Errorf("unbalanced brackets, [, and ]")
			}
		case '{':
			openConjugateCount++
			conjugateBalance++
		case '}':
			closeConjugateCount++
			conjugateBalance--
			if conjugateBalance < 0 {
				return Node{}, fmt.Errorf("unbalanced conjugate, {, and }")
			}
		}
	}
	if openParenCount != closeParenCount {
		return Node{}, fmt.Errorf("unbalanced parentheses, (, and )")
	}
	if openBracketCount != closeBracketCount {
		return Node{}, fmt.Errorf("unbalanced brackets, [, and 	]")
	}
	if openConjugateCount != closeConjugateCount {
		return Node{}, fmt.Errorf("unbalanced conjugate, {, and }")
	}

	stack := [][]Node{{}}
	nstack := make([]bool, 0)
	wasNegated := false
	rstack := make([]string, 0)
	reflectAll := ""
	for i := 0; i < len(input); i++ {
		char := input[i]

		switch char {
		case 'x', 'y', 'z', 'w':
			// modify stack top to not be ""
			if len(rstack) == 0 {
				reflectAll = string(char)
			} else {
				if len(stack[len(stack)-1]) > 0 {
					return Node{}, fmt.Errorf("reflection must first item in the group: %c", char)
				}
				rstack[len(rstack)-1] = string(char)
			}
		case '(', '[', '{':
			stack = append(
				stack,
				[]Node{},
			)
			nstack = append(nstack, wasNegated)
			rstack = append(rstack, "")
		case ')', ']', '}':
			if len(stack) > 1 {
				if char == '}' || char == ']' {
					if len(stack[len(stack)-1]) > 2 {
						return Node{}, fmt.Errorf("use parenthesis to make %c use 2 items, to keep it from being ambiguous", char)
					}
				}
				negatedParens := nstack[len(nstack)-1]
				nstack = nstack[:len(nstack)-1]

				reflection := rstack[len(rstack)-1]
				rstack = rstack[:len(rstack)-1]

				last := stack[len(stack)-1]
				stack = stack[:len(stack)-1]

				stack[len(stack)-1] = append(
					stack[len(stack)-1],
					Node{
						Commutator: char == ']' || char == '}',
						Conjugated: char == '}',
						Arr:        last,
						Negate:     negatedParens,
						Repeat:     1, // maybe updted
						Reflection: reflection,
					},
				)
			}
		case '/':
			// use it to set negate on next token. literal // is ignored.
			wasNegated = !wasNegated
			continue
		case 'U', 'R', 'F', 'D', 'L', 'B', 'u', 'r', 'f', 'd', 'l', 'b':
			face := char
			top := len(stack) - 1
			stack[top] = append(
				stack[top],
				Node{
					Face:   string(face),
					Negate: wasNegated,
					Repeat: 1, // maybe update
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
		default:
			// why do spaces make it stop? return nothing if it wont be interpreted right.
			return Node{}, fmt.Errorf("unexpected character: %c", char)
		}
		// don't skip unless you set these
		wasNegated = false
	}
	popped := stack[len(stack)-1]
	stack = stack[:len(stack)-1]
	n := Node{
		Arr:        popped,
		Reflection: reflectAll,
	}
	return n, nil
}

func (cube *Cube) Pop() bool {
	if len(cube.History) == 0 {
		return false
	}
	// write stickers from history over current stickers
	for k, v := range cube.History[len(cube.History)-1] {
		cube.Stickers[k] = v
	}
	// remove the last history
	cube.History = cube.History[:len(cube.History)-1]
	return true
}

func (cube *Cube) ExecuteCommand(node Node) (string, error) {
	// append a copy of the stickers before this execution
	cube.History = append(cube.History, make(map[string]string))
	for k, v := range cube.Stickers {
		cube.History[len(cube.History)-1][k] = v
	}
	return cube.Execute(node, 0, 0, 0, 0, 0)
}

func (cube *Cube) Execute(node Node, negates, xflips, yflips, zflips, wflips int) (string, error) {
	outcome := ""
	repeat := 1
	// globally track repeats we are under
	if node.Repeat != 0 {
		repeat = node.Repeat
	}
	// globally track the number of negates we are under
	if node.Negate {
		negates++
	}
	if node.Arr != nil {
		if node.Reflection != "" {
			switch node.Reflection {
			case "x":
				xflips++
			case "y":
				yflips++
			case "z":
				zflips++
			case "w":
				wflips++
			}
		}
		fwd := make([]Node, 0)
		for i := 0; i < len(node.Arr); i++ {
			n := node.Arr[i]
			if negates%2 == 1 {
				n = node.Arr[len(node.Arr)-1-i]
			}
			fwd = append(fwd, n)
		}
		// interpret as repeats bind latest
		for i := 0; i < repeat; i++ {
			if !node.Commutator || (node.Commutator && negates%2 == 0) {
				for _, cmd := range fwd {
					result, err := cube.Execute(cmd, negates, xflips, yflips, zflips, wflips)
					if err != nil {
						return outcome, fmt.Errorf("error in %s at %s: %s", outcome, result, err)
					}
					outcome += result
				}
			}
			if node.Commutator {
				for i := 0; i < len(fwd); i++ {
					cmd := fwd[i]
					if !node.Conjugated || (i == 0 && negates%2 == 0) || (i != 0 && negates%2 == 1) {
						result, err := cube.Execute(cmd, negates+1, xflips, yflips, zflips, wflips)
						if err != nil {
							return outcome, fmt.Errorf("error in %s at %s: %s", outcome, result, err)
						}
						outcome += result
					}
				}
				if negates%2 == 1 {
					for _, cmd := range fwd {
						result, err := cube.Execute(cmd, negates, xflips, yflips, zflips, wflips)
						if err != nil {
							return outcome, fmt.Errorf("error in %s at %s: %s", outcome, result, err)
						}
						outcome += result
					}
				}
			}
		}
	} else {
		negates += xflips
		negates += yflips
		negates += zflips
		negates += wflips
		facemap := map[string]string{
			"u": "u",
			"r": "r",
			"f": "f",
			"d": "d",
			"l": "l",
			"b": "b",
			"U": "U",
			"R": "R",
			"F": "F",
			"D": "D",
			"L": "L",
			"B": "B",
		}
		if xflips%2 == 1 {
			facemap["r"] = "l"
			facemap["l"] = "r"
			facemap["R"] = "L"
			facemap["L"] = "R"
		}
		if yflips%2 == 1 {
			facemap["u"] = "d"
			facemap["d"] = "u"
			facemap["U"] = "D"
			facemap["D"] = "U"
		}
		if zflips%2 == 1 {
			facemap["f"] = "b"
			facemap["b"] = "f"
			facemap["F"] = "B"
			facemap["B"] = "F"
		}
		f := facemap[node.Face]
		if node.Face != "" {
			turn := repeat
			turn = turn * (1 - 2*(negates%2))
			rstr := ""
			if repeat != 1 {
				rstr = fmt.Sprintf("%d", repeat)
			}
			if negates%2 == 0 {
				outcome += fmt.Sprintf("%s%s ", f, rstr)
			} else {
				outcome += fmt.Sprintf("/%s%s ", f, rstr)
			}
			cube.Turn(f, turn)
		}
	}
	return outcome, nil
}

// print a message in red if ansi
func (cube *Cube) PrintRed(msg string) {
	if UseAnsi {
		fmt.Printf("\u001b[1;31m%s\u001b[0m\n", msg)
	} else {
		fmt.Printf("%s\n", msg)
	}
}

func Loop() {
	cube := NewCube()
	cube2 := NewCube()

	// loop to get and anlyze a line and draw the screen
	cmd := ""
	repeats := 0
	prevCmd := ""
	cube.Help()

	rdr := bufio.NewReader(os.Stdin)
	for {
		Draw(cmd, repeats, []*Cube{cube, cube2})

		fmt.Printf("\u25B6 ")
		var err error
		cmd, err = rdr.ReadString('\n')
		if err != nil {
			fmt.Printf("Error reading input\n")
			continue
		}
		cmd = strings.TrimSpace(cmd)

		if cmd == "a" {
			UseAnsi = !UseAnsi
			continue
		}

		if cmd == "q" || cmd == "quit" || cmd == "exit" {
			break
		}

		if cmd == "test" {
			cube.PostTest()
			continue
		}

		if cmd == "?" || cmd == "h" || cmd == "help" {
			cube.Help()
			rdr.ReadString('\n')
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

		if cmd == "p" {
			didPop := cube.Pop()
			if !didPop {
				cube.PrintRed("nothing to undo!\n")
			}
			fmt.Printf("stack size: %d\n", len(cube.History))
			continue
		}

		if len(cmd) > 0 && cmd[0] == 'n' {
			cube = NewCube()
			cmd = cmd[1:]
		}

		if cmd == "n" {
			cube = NewCube()
			repeats = 0
			continue
		}

		if cmd == "s" {
			cube, cube2 = cube2, cube
			repeats = 0
			continue
		}

		nodes, err := cube.Parse(cmd)
		if err != nil {
			cube.Help()
			msg := fmt.Sprintf("parse error. see help above: %s\n", err)
			cube.PrintRed(msg)
			continue
		}

		fmt.Printf("parsed as: %s\n", nodes.Print())
		flattened, err := cube.ExecuteCommand(nodes)
		if err != nil {
			cube.Help()
			msg := fmt.Sprintf("execute error. see help above: %s\n", err)
			cube.PrintRed(msg)
			continue
		}
		fmt.Printf("executed moves: %s\n", flattened)
		fmt.Println()
		fmt.Println()
	}
}

func main() {
	flag.Parse()
	if *PostTest {
		cube := NewCube()
		cube.PostTest()
	} else {
		Loop()
	}
}
