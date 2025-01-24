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
		tmp := cube.Stickers[a]
		cube.Stickers[a] = cube.Stickers[b]
		cube.Stickers[b] = tmp
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
		i := cube.Adj[f][fi]                     //at fi
		j := cube.Adj[f][(fi+1)%cube.FacePeriod] //ahead fi
		e0a := s(f + i)
		e1a := s(i + f)
		e0b := s(f + j)
		e1b := s(j + f)
		swap(e0a, e0b)
		swap(e1a, e1b)
		fmt.Printf("%s%s %s%s\n", e0a, e1a, e0b, e1b)

		c0a := s(f + i + k)
		c1a := s(i + k + f)
		c2a := s(k + f + i)
		c0b := s(f + j + i)
		c1b := s(j + i + f)
		c2b := s(i + f + j)
		swap(c0a, c0b)
		swap(c1a, c1b)
		swap(c2a, c2b)
		fmt.Printf("%s%s%s %s%s%s\n", c0a, c1a, c2a, c0b, c1b, c2b)

		if center {
			m0a := s(i)
			m0b := s(j)
			swap(m0a, m0b)
			fmt.Printf("%s %s\n", m0a, m0b)
			e0a := s(i + k)
			e1a := s(k + i)
			e0b := s(j + i)
			e1b := s(i + j)
			swap(e0a, e0b)
			swap(e1a, e1b)
			fmt.Printf("%s%s %s%s\n", e0a, e1a, e0b, e1b)
		}
	}
}

// turn a face *count* times, all cube or just a face
func (cube *Cube) Turn(i string, count int, all bool) {
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

func (cube *Cube) Loop() {
	// loop to get and anlyze a line and draw the screen
	cmd := ""
	repeats := 0
	prevCmd := ""
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
			fmt.Printf("stickers: %v\n", cube.Stickers)
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

		// loop to parse cmd and do what cmd says
		i := 0
		for {
			// nothing to read
			if i >= len(cmd) {
				break
			}

			// skip whitespace
			for cmd[i] == ' ' && i+1 < len(cmd) {
				i++
			}

			// remember if next token is negative
			negative := false
			if cmd[i] == '/' && i+1 < len(cmd) {
				negative = true
				i++
			}

			// expect groups or turn tokens
			// TODO: ( ) [ ]

			// from here, everything is upper or lower-case face.
			c := string(cmd[i])
			toLower := strings.ToLower(c)
			isLower := true
			if strings.Compare(c, toLower) != 0 {
				isLower = false
			}

			// this is about an individual face
			if toLower == "u" || toLower == "r" || toLower == "f" || toLower == "d" || toLower == "l" || toLower == "b" {
				if !isLower {
					cube.Turn(toLower, repeats, negative)
				} else {
					cube.Turn(toLower, repeats, negative)
				}
			}
			if i+1 >= len(cmd) {
				break
			}
			i = i + 1
		}
	}
}

func main() {
	cube := NewCube()
	cube.Loop()
}
