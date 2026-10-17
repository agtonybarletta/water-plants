package main

import (
	"math/rand"
)

type GameState int

const (
	StatePlaying GameState = iota
	StateWon
	StateLost
)

type Cell struct {
	wet     bool
	digged  bool
	planted bool
}
type Game struct {
	horTiles   int // number of rows
	verTiles   int // number of columns
	cells      [][]Cell
	currRow    int
	currCol    int
	shovels    int
	plantsLeft int
	state      GameState
}

const SHOVELS_COMPUTED = -1

func (g *Game) init(horTiles, verTiles, shovels int) {
	g.horTiles = horTiles
	g.verTiles = verTiles
	g.cells = make([][]Cell, horTiles)
	for i := 0; i < horTiles; i++ {
		g.cells[i] = make([]Cell, verTiles)
	}
	g.cells[g.horTiles-1][g.verTiles-1].digged = true
	g.currRow = g.horTiles - 1
	g.currCol = g.verTiles - 1
	if shovels == SHOVELS_COMPUTED {
		g.shovels = g.horTiles + g.verTiles // TODO compute the right amount
	} else {
		g.shovels = shovels
	}
	g.state = StatePlaying
}

func (g *Game) Init(horTiles, verTiles, shovels int, plantPositions [][2]int) {
	g.init(horTiles, verTiles, shovels)
	g.plantsLeft = len(plantPositions)
	for _, p := range plantPositions {
		row, col := p[0], p[1]
		if row >= 0 && row < horTiles && col >= 0 && col < verTiles {
			g.cells[row][col].planted = true
		}
	}
	g.wetNeighborCells(g.currRow, g.currCol)
}

/* Given a game (grid of cells)
 * It compute the minimum number of shovels
 * needed to win the game
 */
func (g *Game) minShovels() int {

	return 99
}

func (game *Game) wetCell(row, col int) {

	if game.state != StatePlaying {
		return
	}
	if row < 0 || row >= game.horTiles || col < 0 || col >= game.verTiles {
		return
	} else {
		if !game.cells[row][col].digged && !game.cells[row][col].wet {
			game.cells[row][col].wet = true
			if game.cells[row][col].planted {
				game.plantsLeft -= 1
				if game.plantsLeft == 0 {
					game.state = StateWon
				}
			}
		}
	}

}

func (game *Game) wetNeighborCells(row, col int) {
	d := []int{-1, 0, 1}

	for i := 0; i < 3; i++ {
		for j := 0; j < 3; j++ {
			if d[i] == 0 && d[j] == 0 {
				continue
			}
			nr := row + d[i]
			nc := col + d[j]
			game.wetCell(nr, nc)
		}
	}

}

func (game *Game) digCell(row, col int) {

	if game.state != StatePlaying {
		return
	}
	if row < 0 || row >= game.horTiles || col < 0 || col >= game.verTiles {
		return
	}
	if !game.cells[row][col].digged {
		game.shovels -= 1
		game.cells[row][col].digged = true
		game.cells[row][col].wet = false
	}
}

func (game *Game) updateState() {
	if game.shovels == 0 && game.plantsLeft > 0 {
		game.state = StateLost
	}
}

func (m *Model) MoveLeft() {
	game := &m.levels[m.currentLevel].game
	if game.currCol > 0 {
		if !game.cells[game.currRow][game.currCol-1].planted {
			game.currCol--
			game.digCell(game.currRow, game.currCol)
			game.wetNeighborCells(game.currRow, game.currCol)
			game.updateState()

		}
	}
}

func (m *Model) MoveRight() {
	game := &m.levels[m.currentLevel].game
	if game.currCol < game.verTiles-1 {
		row, col := game.currRow, game.currCol+1
		if !game.cells[row][col].planted {
			game.currCol++
			game.digCell(game.currRow, game.currCol)
			game.wetNeighborCells(game.currRow, game.currCol)
			game.updateState()
		}
	}
}

func (m *Model) MoveUp() {
	game := &m.levels[m.currentLevel].game
	if game.currRow > 0 {
		row, col := game.currRow-1, game.currCol
		if !game.cells[row][col].planted {
			game.currRow--
			game.digCell(game.currRow, game.currCol)
			game.wetNeighborCells(game.currRow, game.currCol)
			game.updateState()
		}
	}
}

func (m *Model) MoveDown() {
	game := &m.levels[m.currentLevel].game
	if game.currRow < game.horTiles-1 {
		row, col := game.currRow+1, game.currCol
		if !game.cells[row][col].planted {
			game.currRow++
			game.digCell(game.currRow, game.currCol)
			game.wetNeighborCells(game.currRow, game.currCol)
			game.updateState()
		}
	}
}

type Level struct {
	level  int
	game   Game
	seed   int
	def    LevelDef
	plants [][2]int
}

func randomPlants(horTiles, verTiles int, rng *rand.Rand) [][2]int {
	totalCells := horTiles * verTiles
	numPlants := totalCells/8 + 1
	if numPlants > totalCells-1 {
		numPlants = totalCells - 1
	}
	// Collect all positions except the start position (bottom-right)
	startRow, startCol := horTiles-1, verTiles-1
	type pos struct{ row, col int }
	positions := make([]pos, 0, totalCells-1)
	for i := 0; i < horTiles; i++ {
		for j := 0; j < verTiles; j++ {
			if i == startRow && j == startCol {
				continue
			}
			positions = append(positions, pos{i, j})
		}
	}
	rng.Shuffle(len(positions), func(i, j int) {
		positions[i], positions[j] = positions[j], positions[i]
	})
	result := make([][2]int, numPlants)
	for k := 0; k < numPlants; k++ {
		result[k] = [2]int{positions[k].row, positions[k].col}
	}
	return result
}

func (l *Level) Init(def LevelDef) {
	l.level = def.level
	var plantPositions [][2]int
	if def.plants != nil {
		plantPositions = def.plants
		l.seed = 0
	} else {
		l.seed = rand.Int()
		rng := rand.New(rand.NewSource(int64(l.seed)))
		plantPositions = randomPlants(def.horTiles, def.verTiles, rng)
	}
	l.plants = plantPositions
	l.game.Init(def.horTiles, def.verTiles, def.shovels, plantPositions)
	l.def = def
}

func (l *Level) ResetFromSeed() {
	l.game.Init(l.game.horTiles, l.game.verTiles, l.def.shovels, l.plants)
}

type LevelDef struct {
	level    int
	horTiles int
	verTiles int
	shovels  int
	plants   [][2]int // nil means random plants
}

var levelDefs = []LevelDef{
	//{level: 1, horTiles: 5, verTiles: 5, shovels: 99, plants: [][2]int{{0, 0}}},
	//{level: 2, horTiles: 5, verTiles: 5, shovels: 99, plants: [][2]int{{0, 0}, {0, 4}}},
	// {level: 3, horTiles: 5, verTiles: 5, shovels: 6, plants: [][2]int{{1, 3}, {4, 0}, {0, 3}}},
	// {level: 4, horTiles: 6, verTiles: 6, shovels: 10, plants: [][2]int{{0, 0}, {0, 2}, {0, 5}, {4, 0}, {4, 2}}},
	// {level: 5, horTiles: 6, verTiles: 6, shovels: SHOVELS_COMPUTED, plants: [][2]int{{0, 0}, {0, 5}, {5, 0}, {2, 2}}},
	{level: 6, horTiles: 7, verTiles: 3, shovels: SHOVELS_COMPUTED, plants: nil},
}

type Model struct {
	levels       []Level
	currentLevel int
}

func (m *Model) Init() {
	m.levels = make([]Level, len(levelDefs))
	for i, def := range levelDefs {
		m.levels[i].Init(def)
	}
	m.currentLevel = 0
}

func (m *Model) GetCurrentLevelNumber() int {
	return m.levels[m.currentLevel].level
}

func (m *Model) GetCurrentLevel() [][]Cell {
	return m.levels[m.currentLevel].game.cells
}

func (m *Model) GetCurrentTilesSize() (int, int) {
	return m.levels[m.currentLevel].game.horTiles, m.levels[m.currentLevel].game.verTiles
}

func (m *Model) GetCurrentPosition() (int, int) {
	return m.levels[m.currentLevel].game.currRow, m.levels[m.currentLevel].game.currCol
}

func (m *Model) GetCurrentPlantsLeft() int {
	return m.levels[m.currentLevel].game.plantsLeft
}

func (m *Model) GetCurrentShovels() int {
	return m.levels[m.currentLevel].game.shovels
}

func (m *Model) GetState() GameState {
	return m.levels[m.currentLevel].game.state
}

func (m *Model) ResetFromSeed() {
	m.levels[m.currentLevel].ResetFromSeed()
}

/* generic move, give up for now
func (m *Model) move(direction int) {
	if direction < 0 || direction > 3 {
		return
	}

	game := &m.levels[m.currentLevel].game
	if game.currY <= 0 || game.currX <= 0 || game.curr
	dx := []int{-1, 0, 1, 0}
	dy := []int{0, -1, 0, 1}

	if (game.currY > 0) || () {
		if !game.cells[game.currX][game.currY-1].planted {
			game.currY--
		}
	}
}
*/

func (m *Model) NextLevel() bool {
	lastIdx := len(m.levels) - 1
	if m.currentLevel < lastIdx {
		m.currentLevel++
	} else {
		// Generate a new endless level based on the current level's dimensions
		lastLevel := &m.levels[m.currentLevel]
		nextLevel := lastLevel.level + 1
		horTiles := lastLevel.game.horTiles
		verTiles := lastLevel.game.verTiles
		newDef := LevelDef{
			level:    nextLevel,
			horTiles: horTiles,
			verTiles: verTiles,
			plants:   nil, // always random
		}
		var level Level
		level.Init(newDef)
		m.levels = append(m.levels, level)
		m.currentLevel++
	}
	return true
}

func (m *Model) Reset() {
	if m.currentLevel < len(levelDefs) {
		m.levels[m.currentLevel].Init(levelDefs[m.currentLevel])
	} else {
		m.ResetFromSeed()
	}
}
