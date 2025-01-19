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

The main thing is that upper-case to move the whole cube by its face,
or use lower case to turn a face, and use "/" to negate a turn.
When you type in a move like "r u /r /u", you can figure out its period,
by pressing return until the cube is solved again. The repeats will
tell you that "r u /r /u" has period 6.

You can use the CLI to solve a cube after you scramble it.
