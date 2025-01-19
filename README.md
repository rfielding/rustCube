rustcube
========

A command-line tool for Rubiks Cube calculations.
We want to draw the state of a rubiks cube like this:

```
  ... bbb ...
 
.     uuu     .
.     uuu     .
.     uuu     .

b lll fff rrr b
b lll fff rrr b
b lll fff rrr b

.     ddd     .
.     ddd     .
.     ddd     .

  ... bbb ...
```

We need to be able to lookup the sticker color
for 1d centers, 2d edges, and 3d corners.
ie, where all corners are clockwise-winding:

  - u: u
  - ur: r
  - rf: r
  - fur: f

All faces enumerate adjacencies as counter-clockwise-winding:

  - f: u l d r
  - u: b l f r

Where it can't quite be forced which adjacent face comes first.
Symmetries just demand that all faces mention in a counter-clockwise
circular order. To twist a face, all involved stickers need to be moved:

  - everything is (period-1) swaps moving counter-clockwise to do a clockwise turn
    - edge on both faces
    - corner on all three faces

Ex of turn f:
   - for i in 0..3: # adjacent[f] = [u, l, d, r]
     - swap (f, adjacent[i])

> assume that a number BEFORE face moves a slice. a number after a slice is the
number of turns. to twist the entire 3x3x3 cube from face u:  u, 1u, d3. we can use / as a shorthand for 3 turns, to get inverse:  r /u /r u could be a commutator about u.

