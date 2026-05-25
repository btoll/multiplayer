package main

type rules interface {
	boundsCheck(position, *dimension) bool
	calculateTarget(position, direction, dimension) position
	hasCollision([][]int, position, int) (int, bool)
	onCollision(int, int) (bool, bool)
}

type biplaneRules struct{}

func (r *biplaneRules) boundsCheck(pos position, dim dimension) bool {
	return true
}

func (r *biplaneRules) calculateTarget(pos position, dir direction, dim *dimension) position {
	return position{
		row: ((pos.row+dir.y)%dim.rows + dim.rows) % dim.rows,
		col: ((pos.col+dir.x)%dim.cols + dim.cols) % dim.cols,
	}
}

func (r *biplaneRules) hasCollision(board [][]int, pos position, pieceID int) (int, bool) {
	cell := board[pos.row][pos.col]
	if cell != pieceID && cell != 0 {
		return cell, true
	}
	return 0, false
}

func (r *biplaneRules) onCollision(a, b int) (bool, bool) {
	return true, true
}

type bulletRules struct{}

func (r *bulletRules) boundsCheck(pos position, dim dimension) bool {
	if pos.row >= 0 && pos.row < dim.rows &&
		pos.col >= 0 && pos.col < dim.cols {
		return true
	}
	return false
}

func (r *bulletRules) calculateTarget(pos position, dir direction, dim *dimension) position {
	return position{
		row: pos.row + dir.y,
		col: pos.col + dir.x,
	}
}

func (r *bulletRules) hasCollision(board [][]int, pos position, pieceID int) (int, bool) {
	cell := board[pos.row][pos.col]
	if cell != pieceID && cell != 0 {
		return cell, true
	}
	return 0, false
}

func (r *bulletRules) onCollision(a, b int) (bool, bool) {
	return true, true
}
