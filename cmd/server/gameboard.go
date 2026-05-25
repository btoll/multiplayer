package main

import (
	"fmt"
	"maps"
	"strings"
	"sync"
)

type color int

const (
	_ color = iota
	RED
	GREEN
	YELLOW
	BLUE
	PURPLE
	_
	WHITE
)

func colorGenerator() func() (color, bool) {
	colors := []color{RED, GREEN, YELLOW, BLUE, PURPLE, WHITE}
	var i int
	return func() (color, bool) {
		if i < len(colors) {
			c := colors[i]
			i += 1
			return c, true
		}
		return -1, false
	}
}

func (c color) String() string {
	return [...]string{
		"_",
		"\033[31m",
		"\033[32m",
		"\033[33m",
		"\033[34m",
		"\033[35m",
		"_",
		"\033[37m",
	}[c]
}

type gameboard interface {
	draw()
	getBoard() [][]int
	getDimension() dimension
	getGamepiece(int) gamepiece
	move(msg)
	register(createGamepiece)
	render() string
	unregister(int)
}
type gamePieces map[int]gamepiece

type biplanes struct {
	board     [][]int
	dimension dimension
	pieces    gamePieces
	mu        sync.RWMutex
}

type dimension struct {
	rows int
	cols int
}

func newGameboard(visibleRows, visibleCols int) gameboard {
	board := make([][]int, visibleRows)
	for i := range board {
		board[i] = make([]int, visibleCols)
	}
	return &biplanes{
		board:     board,
		pieces:    make(gamePieces),
		dimension: dimension{visibleRows, visibleCols},
	}
}

func (b *biplanes) draw() {
	b.mu.Lock()
	pieces := make(gamePieces, len(b.pieces))
	maps.Copy(pieces, b.pieces)
	b.mu.Unlock()

	for _, gamepiece := range pieces {
		isStopped := gamepiece.isStopped()
		if isStopped {
			continue
		}
		gamepiece.draw(b)
	}
}

func (b *biplanes) getBoard() [][]int {
	return b.board
}

func (b *biplanes) getDimension() dimension {
	return b.dimension
}

func (b *biplanes) getGamepiece(id int) gamepiece {
	return b.pieces[id]
}

func (b *biplanes) move(m msg) {
	piece := b.pieces[m.id]
	if piece == nil {
		return
	}
	b.mu.Lock()
	piece.move(m.x, m.y)
	b.mu.Unlock()
}

func (b *biplanes) register(c createGamepiece) {
	b.mu.Lock()
	piece := newGamePiece(c)
	// Put the piece on the board.
	// The id is looked up when rendering and the cell colorized.
	// TODO
	switch v := piece.(type) {
	case *biplane:
		b.board[v.cur.row][v.cur.col] = c.id
	case *bullet:
		b.board[v.cur.row][v.cur.col] = c.id
	}
	b.pieces[c.id] = piece
	b.mu.Unlock()
}

func (b *biplanes) render() string {
	b.mu.Lock()
	rows := len(b.board)
	cols := 0
	if rows > 0 {
		cols = len(b.board[0])
	}
	cpBoard := make([][]int, rows)
	for row := range cpBoard {
		cpBoard[row] = make([]int, cols)
		copy(cpBoard[row], b.board[row])
	}
	b.mu.Unlock()

	builder := &strings.Builder{}
	builder.WriteString("\033[2J\033[H")
	for r, row := range cpBoard {
		for c, cell := range row {
			if cell == 0 {
				builder.WriteByte(' ')
			} else {
				piece := b.pieces[cell]
				// CRITICAL: Handle the case where unregister left a ghost ID
				if piece == nil {
					fmt.Println("got here", cell)
					// 1. Don't crash: treat as empty
					builder.WriteByte(' ')
					// 2. CLEAN UP: Remove the ghost ID from the REAL board
					// We must lock again to safely modify the board
					b.mu.Lock()
					// Double-check: maybe another thread cleared it already?
					if b.board[r][c] == cell {
						b.board[r][c] = 0
					}
					b.mu.Unlock()
					continue
				}
				fmt.Fprintf(builder, "%s%c\033[0m", piece.getColor(), piece.getSymbol())
			}
		}
		builder.WriteByte('\n')
	}
	return builder.String()
}

func (b *biplanes) unregister(id int) {
	b.mu.Lock()
	defer b.mu.Unlock()
	// Clear the board slot BEFORE deleting from map
	// This ensures render() never sees a mismatch
	piece := b.getGamepiece(id)
	if piece == nil {
		return
	}
	pos := piece.getPosition()
	b.board[pos.row][pos.col] = 0
	delete(b.pieces, id)
	fmt.Printf("[INFO] User ID %d unregistered.\n", id)
}
