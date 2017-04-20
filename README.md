
Minesweeper-solver
==================

This is an attempt at a solver for minesweeper. It is a client for
[DefuseDivision](https://github.com/lelandbatey/defuse_division), a minesweeper
server.

The solver uses probability and, (if not 100% sure of which cell to probe)
linear algebra in its calculations. The solver moves from top to bottom, left
to right. It first flags all mines it knows about, then moves on to probing
known non-mine cells (safe cells). So long as it continues to find mines to
flag and safe cells to probe using probability, it will refrain from using the
results from its linear algebra calculations. This can lead the solver to
ignore portions of the minefield that is unsolveable using probability. You'll
see this behavior when the solver begins probing below some rows of unprobed
cells. As more information is revealed, these ignored portions often become
solveable and so are accordingly flagged and probed.  
___  
Once forced to use linear algebra, the solver begins by flagging mines made
visible by linear algebra, then moving on to probing safe cells.  If, even
after using linear algebra, it is still not 100% certain, the solver will pick
the best solution (based on probability) and probe it. It will indicate its
uncertainty by circling around the cell before probing.

## Some Terminology
I'll be using these terms a lot, so there's two important terms to learn:
#### Witness
A witness is a cell that's been probed (aka revealed) and is nearby a mine.
You'll see this on the board as a number. A witness with a value of 4 means
that there are 4 mines neighboring this witness.
#### Primed Cell
A primed cell is an unprobed cell bordering a witness. When playing
minesweeper, these are almost exclusively the cells you'll click because you
have information (somewhat) about their contents. We call them primed because
they could (or could not) be a mine. So they are "Primed" to explode.

## So How Does This Work?
#### Using Probability
This solver uses probability to determine the presence of mines and safe cells.
It asks each witness two simple questions:
 1. How many mines are nearby?
 2. How many primed cells are nearby?

If mines == primed, then each of those neighboring primed cells has a 100%
chance to be a mine. If there are 2 mines nearby, and 4 primed nearby, then
each of those neighboring primed has a 50% chance of being a mine. You see
where this is going:  
**% chance of primed neighbor being a mine = (# of mines / # of primed neighbors)**  
A primed-cell's % probability is an average of its probability
assigned by each of its neighboring witness cells. Of course, there's some
catches to this. As soon as we deduce a mine or safe cell, we must recalculate
the probability of the entire board, taking into account known-state cells.
Witness answers to the questions about mine-count and primed-count are affected
by these known-state cells.
#### Using Linear Algebra
Linear Algebra solves minesweeper by realizing that each witness (and its
primed neighbors) can be represented as an equation. A witness with a value of
1, bording on cells C1 and C2 can be represented as: `C1 + C2 = 1`, where the
values of C1 & C2 are 1 or 0 (representing the presence or absence of a mine,
respectively). An adjacent witness with value of 2, bording cells C1, C2, and
C3 can be represented as: `C1 + C2 + C3 = 2`. If we set these two equations
next to each other like so:  
`C1 + C2 + 0  = 1`  
`C1 + C2 + C3 = 2`  
you can see that we can subtract equation 1 from equation 2, thus ending up
with `C3 = 1` which tells us immediately that C3 is a mine!  
There is an automated way to do this with an equation for every witness. First
we convert each witness-equation into a sequence of numbers representing the
coefficients in its equation. This sequence becomes a row in a matrix. The full
matrix has a row for each witness in the minefield. For example, if we were
dealing with just the two previously mentioned witnesses, their corresponding
equations would be converted into this matrix:  
1 | 1 | 0 | 1
---|---|---|---
1 | 1 | 1 | 2
From here, we perform row operations (scaling, adding, subtracting) to arrive
at a **Row-Reduced Echelon Form**, which will reduce the above matrix to:  
1  | 1 | -1 | 0
---|---|---|---
0  | 0 | 1 | 1
which is equivalent to the following two equations:  
`C1 + C2 - C3 = 0` and `C3 = 1`  
Again, we see that C3 is a mine. The first equation, however, remains
ambiguous. Is C1 or C2 a mine? One of them is... To solve this, we will need
more info. More equations. A larger matrix. Luckily, Row-Reduced Echelon Form
is incredibly fast to compute, which is why this solver calculates RREF on
every step. Treating a minefield in this fashion can puzzle out any solvable
scenario. (It should be noted that minesweeper is not solveable 100% of the
time)
# Give it a whirl
You'll need to download and run the DefuseDivision server  
`git clone https://github.com/lelandbatey/defuse_division`  
`cd defuse_division`  
`python3 play-defusedivision`  
_select **Host and Play**_  
_make sure it's running in a large / fullscreen terminal window_  

Then download and run minesweeper-solver in another terminal:  
`go get github.com/lelandbatey/minesweeper-solver`  
`cd $GOPATH/src/github.com/lelandbatey/minesweeper-solver`  
`go run main.go`  
_Watch minesweeper get schooled_
