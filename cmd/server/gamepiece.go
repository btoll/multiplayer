package main

import "fmt"

type gamepiece interface {
	draw(gameboard)
	getColor() color
	getCardinalDirection() string
	getDirection() direction
	getPosition() position
	getSymbol() rune
	isStopped() bool
	move(int, int)
	setDirection(direction)
}

type direction struct {
	x int
	y int
}

type position struct {
	row int
	col int
}

type pieceBase struct {
	id     int
	color  color
	symbol rune
	prev   position
	cur    position
	dir    direction
}

func (g pieceBase) getColor() color {
	return g.color
}

func (g pieceBase) getCardinalDirection() string {
	switch g.dir {
	case direction{0, 1}:
		return "north"
	case direction{1, 0}:
		return "east"
	case direction{0, -1}:
		return "south"
	case direction{-1, 0}:
		return "west"
	default:
		return "TODO"
	}
}

func (g pieceBase) getDirection() direction {
	return g.dir
}

func (g pieceBase) getPosition() position {
	return g.cur
}

func (g pieceBase) getSymbol() rune {
	return g.symbol
}

func (g pieceBase) isStopped() bool {
	return g.dir.x == 0 && g.dir.y == 0
}

func (g *pieceBase) move(x, y int) {
	g.dir = direction{x, y}
}

func (g *pieceBase) setDirection(d direction) {
	g.dir = d
}

func newGamePiece(c createGamepiece) gamepiece {
	switch c.kind {
	case "biplane":
		return newBiplane(c)
	case "bullet":
		return newBullet(c)
	}
	return nil
}

// Compile-time assertion, ensure that *biplane implements gamepiece without constructing
// a value (no allocation).  Checks method-set compatibililty for *biplane.
// var _ Interface = (*T)(nil)
var _ gamepiece = (*biplane)(nil)

type biplane struct {
	*pieceBase
	rules biplaneRules
}

func newBiplane(c createGamepiece) *biplane {
	return &biplane{
		pieceBase: &pieceBase{
			id:    c.id,
			color: c.color,
			//			symbol: '🚁',
			symbol: 'X',
			cur:    c.position,
		},
		rules: biplaneRules{},
	}
}

func (b *biplane) draw(gameboard gameboard) {
	board := gameboard.getBoard()
	dim := gameboard.getDimension()
	pos := b.rules.calculateTarget(b.cur, b.dir, &dim)
	victimID, collided := b.rules.hasCollision(board, pos, b.id)
	if collided {
		m, n := b.rules.onCollision(b.id, victimID)
		if m {
			gameboard.unregister(b.id)
		}
		if n {
			gameboard.unregister(victimID)
		}
		return
	}

	board[b.cur.row][b.cur.col] = 0
	b.prev.row = b.cur.row
	b.prev.col = b.cur.col
	b.cur.row = pos.row
	b.cur.col = pos.col
	board[b.cur.row][b.cur.col] = b.id
}

var _ gamepiece = (*bullet)(nil)

type bullet struct {
	*pieceBase
	rules bulletRules
}

func newBullet(c createGamepiece) *bullet {
	if c.direction.x > 0 {
		c.position.col += 4
	}
	if c.direction.x < 0 {
		c.position.col -= 4
	}
	if c.direction.y > 0 {
		c.position.row += 4
	}
	if c.direction.y < 0 {
		c.position.row -= 4
	}
	//	if c.position.row >= c.dimension.rows ||
	//		c.position.col >= c.dimension.cols {
	//		return nil
	//	}
	fmt.Fprintf(logFile, "new bullet=%+v\n", c)
	return &bullet{
		pieceBase: &pieceBase{
			id:     c.id,
			color:  c.color,
			symbol: '-',
			cur:    c.position,
			dir: direction{
				x: c.direction.x,
				y: c.direction.y,
			},
		},
		rules: bulletRules{},
	}
}

func (b *bullet) draw(gameboard gameboard) {
	board := gameboard.getBoard()

	b.prev.row = b.cur.row
	b.prev.col = b.cur.col

	destRow := b.cur.row + b.dir.y
	destCol := b.cur.col + b.dir.x
	pos := position{destRow, destCol}

	// Bounds-check.
	if b.rules.boundsCheck(pos, gameboard.getDimension()) {
		victimID, collided := b.rules.hasCollision(board, pos, b.id)
		// if cell != 0 && cell != b.id {
		if collided {
			// Remove the hit piece.
			m, n := b.rules.onCollision(b.id, victimID)
			if m {
				gameboard.unregister(b.id)
			}
			if n {
				gameboard.unregister(victimID)
			}
			return
		}

		// Move the bullet.
		b.cur.row = destRow
		b.cur.col = destCol

		// No collision: Clear old position, set new position.
		board[b.prev.row][b.prev.col] = 0
		board[b.cur.row][b.cur.col] = b.id
	} else {
		// Out of bounds: Clear both old and current positions.
		board[b.prev.row][b.prev.col] = 0
		board[b.cur.row][b.cur.col] = 0
		// Unregister the bullet.
		gameboard.unregister(b.id)
		return
	}
}
