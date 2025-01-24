package main

import (
	"fmt"
)

/*
  This is a Go implementation just so that I can get it done.
  It is very unproductive to use Rust for writing a parser.
  I will work on the Rust version to learn Rust, but the Go version
  to get the parsing to work right.
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
		// adjacencies
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

// 1 turn of 1 face at i,
//
//	physical parts: ru~ur, rub~ubr~bru
func (cube *Cube) Turn1(f string, center bool) {
	// faces have a period of 4, move their stickers
	//
	//  k | f | j
	//    -----
	//      i
	//
	// edge:   fi,if -> fj,jf
	// corner: fik,ikf,kfi -> fji,jif,fji
	//
	for fi := 0; fi < cube.FacePeriod; fi++ {
		k := cube.Adj[f][(fi+3)%cube.FacePeriod] //behind fi
		i := cube.Adj[f][fi]                     //at fi
		j := cube.Adj[f][(fi+1)%4]               //ahead fi

		e0a := cube.Stickers[f+i]
		e1a := cube.Stickers[i+f]

		e0b := cube.Stickers[f+j]
		e1b := cube.Stickers[j+f]

		c0a := cube.Stickers[f+j+i]
		c1a := cube.Stickers[j+i+f]
		c2a := cube.Stickers[i+f+j]

		c0b := cube.Stickers[f+i+k]
		c1b := cube.Stickers[i+k+f]
		c2b := cube.Stickers[k+f+i]

		// swap a and b, corner and edge orbit
		cube.Stickers[e0b], cube.Stickers[e0a] = cube.Stickers[e0a], cube.Stickers[e0b]
		cube.Stickers[e1b], cube.Stickers[e1a] = cube.Stickers[e1a], cube.Stickers[e1b]
		cube.Stickers[c0b], cube.Stickers[c0a] = cube.Stickers[c0a], cube.Stickers[c0b]
		cube.Stickers[c1b], cube.Stickers[c1a] = cube.Stickers[c1a], cube.Stickers[c1b]
		cube.Stickers[c2b], cube.Stickers[c2a] = cube.Stickers[c2a], cube.Stickers[c2b]

		if center {
			m0a := i
			e0a := i + k
			e1a := k + i

			m0b := j
			e0b := j + i
			e1b := i + j

			cube.Stickers[m0b], cube.Stickers[m0a] = cube.Stickers[m0a], cube.Stickers[m0b]
			cube.Stickers[e0b], cube.Stickers[e0a] = cube.Stickers[e0a], cube.Stickers[e0b]
			cube.Stickers[e1b], cube.Stickers[e1a] = cube.Stickers[e1a], cube.Stickers[e1b]
		}
	}
}

// turn a face *count* times, all cube or just a face
func (cube *Cube) Turn(i string, count int, all bool) {
	// normalize turn count
	for count < 0 {
		count += cube.FacePeriod
	}
	count %= cube.FacePeriod

	for n := 0; n < count; n++ {
		if all {
			// turn an entire 3x3x3 by turning top and center by count, and bottom by -count
			cube.Turn1(i, true)
			cube.Turn1(cube.Opposite[i], false)
		} else {
			// just turn a face
			cube.Turn1(i, false)
		}
	}
}

// stdio side-effect
func (cube *Cube) Draw(cmd string, repeats int) {
	s := func(s string) string {
		// it is impossible to look up a sticker and not find it unless we misspelled the name
		v := cube.Stickers[s]
		if v == "" {
			panic(fmt.Sprintf("sticker is not mapped: %s\n%v", s, cube.Stickers))
		}
		return v
	}

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
	cmd := ""
	repeats := 0
	for {
		fmt.Scanln(&cmd)

		if cmd == "q" {
			break
		}

		if cmd == "n" {
			cube = NewCube()
			continue
		}

		repeats = 1

		_ = cmd
		_ = repeats

		cube.Draw(cmd, repeats)
		fmt.Println()
	}
}

func main() {
	cube := NewCube()
	cube.Loop()
}
