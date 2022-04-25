package main

import (
	"bufio"
	"fmt"
	"math/rand"
	"os"
	"runtime"
	"strconv"
	"strings"
	"time"
)

var Reset = "\033[0m"
var Red = "\033[31m"
var Green = "\033[32m"
var Yellow = "\033[33m"
var Blue = "\033[34m"
var Purple = "\033[35m"
var Cyan = "\033[36m"
var Gray = "\033[37m"
var White = "\033[97m"

func init() {
	if runtime.GOOS == "windows" {
		Reset = ""
		Red = ""
		Green = ""
		Yellow = ""
		Blue = ""
		Purple = ""
		Cyan = ""
		Gray = ""
		White = ""
	}
}

type Space struct {
	IsMine   bool
	IsMarked bool
	IsDug    bool
}

type Board struct {
	width  int
	height int
	mines  int
	spaces map[int]map[int]*Space

	flags     int
	dugSpaces int

	generated bool
}

func NewBoard(width, height, mines int) *Board {
	return &Board{
		width:  width,
		height: height,
		mines:  mines,
		spaces: make(map[int]map[int]*Space),
	}
}

func (b *Board) GenerateMines(firstX, firstY int) {
	openSpaces := genArr(b.width, b.height, firstX, firstY)

	for i := 0; i < b.mines; i++ {
		spaceIndex := rand.Intn(len(openSpaces))
		space := openSpaces[spaceIndex]
		b.AddMine(space.x, space.y)

		openSpaces = append(openSpaces[0:spaceIndex], openSpaces[spaceIndex+1:]...)

	}

	b.generated = true
}

func (b *Board) AddMine(x, y int) {
	if _, ok := b.spaces[x]; !ok {
		b.spaces[x] = make(map[int]*Space)
	}
	if _, ok := b.spaces[x][y]; !ok {
		b.spaces[x][y] = &Space{}
	}
	b.spaces[x][y].IsMine = true
}

func (b *Board) IsMine(x, y int) bool {
	if _, ok := b.spaces[x]; ok {
		if s, ok := b.spaces[x][y]; ok && s.IsMine {
			return true
		}
	}
	return false
}

func (b *Board) IsDug(x, y int) bool {
	if _, ok := b.spaces[x]; ok {
		if s, ok := b.spaces[x][y]; ok && s.IsDug {
			return true
		}
	}
	return false
}

func (b *Board) IsMarked(x, y int) bool {
	if _, ok := b.spaces[x]; ok {
		if s, ok := b.spaces[x][y]; ok && s.IsMarked {
			return true
		}
	}
	return false
}

func (b *Board) Mark(x, y int) {
	if _, ok := b.spaces[x]; !ok {
		b.spaces[x] = make(map[int]*Space)
	}
	if _, ok := b.spaces[x][y]; !ok {
		b.spaces[x][y] = &Space{}
	}
	if b.spaces[x][y].IsMarked {
		b.flags = b.flags - 1
	} else {
		b.flags = b.flags + 1
	}
	b.spaces[x][y].IsMarked = !b.spaces[x][y].IsMarked
}

func (b *Board) Dig(origX, origY int) bool {
	if !b.generated {
		b.GenerateMines(origX, origY)
	}

	if b.IsMarked(origX, origY) {
		return true
	}
	if b.IsMine(origX, origY) {
		return false
	}
	pointsToCheck := []*Point{{x: origX, y: origY}}
	checkedPoints := make(map[int]map[int]bool)
	for len(pointsToCheck) > 0 {
		pt := pointsToCheck[len(pointsToCheck)-1]
		pointsToCheck = pointsToCheck[:len(pointsToCheck)-1]
		if _, ok := checkedPoints[pt.x]; !ok {
			checkedPoints[pt.x] = make(map[int]bool)
		}
		checkedPoints[pt.x][pt.y] = true

		if _, ok := b.spaces[pt.x]; !ok {
			b.spaces[pt.x] = make(map[int]*Space)
		}
		if _, ok := b.spaces[pt.x][pt.y]; !ok {
			b.spaces[pt.x][pt.y] = &Space{}
		}
		if !b.spaces[pt.x][pt.y].IsDug {
			b.dugSpaces = b.dugSpaces + 1
		}
		b.spaces[pt.x][pt.y].IsDug = true

		ct := b.CountAdjacents(pt.x, pt.y)
		if ct > 0 {
			continue
		}
		for i := pt.y - 1; i < pt.y+2; i++ {
			if i < 0 || i > b.height-1 {
				continue
			}
			for j := pt.x - 1; j < pt.x+2; j++ {
				if j < 0 || j > b.width-1 {
					continue
				}
				if _, ok := checkedPoints[j]; ok {
					if _, ok := checkedPoints[j][i]; ok {
						continue
					}
				}
				pointsToCheck = append(pointsToCheck, &Point{
					x: j,
					y: i,
				})
				if _, ok := checkedPoints[j]; !ok {
					checkedPoints[j] = make(map[int]bool)
				}
				checkedPoints[j][i] = true
			}
		}
	}
	return true
}

func (b *Board) CountAdjacents(x, y int) int {
	count := 0
	for i := y - 1; i < y+2; i++ {
		for j := x - 1; j < x+2; j++ {
			if b.IsMine(j, i) {
				count = count + 1
			}
		}
	}
	return count
}

func (b *Board) CheckWin() bool {
	fmt.Println("Flags", b.flags)
	fmt.Println("Mines", b.mines)
	fmt.Println("Dug", b.dugSpaces)
	fmt.Println("Expected Dug", b.width*b.height-b.mines)
	return b.flags == b.mines && (b.dugSpaces == b.width*b.height-b.mines)
}

func (b *Board) String() string {
	ret := ""
	for i := 0; i < b.height; i++ {
		for j := 0; j < b.width; j++ {
			ret = fmt.Sprintf("%s[%s]", ret, func() string {
				if b.IsDug(j, i) {
					ct := b.CountAdjacents(j, i)
					if ct > 0 {
						return fmt.Sprintf("%s%d%s", Yellow, ct, Reset)
					} else {
						return " "
					}
				}
				if b.IsMarked(j, i) {
					return Green + "F" + Reset
				}
				return Blue + "?" + Reset
			}())
		}
		ret += "\n"
	}
	return ret
}

func main() {
	rand.Seed(time.Now().UnixNano())

	width := 20
	height := 20
	mines := (width * height) / 5

	board := NewBoard(width, height, mines)

	for {
		fmt.Println(board)
		fmt.Printf("Flags: %d/%d\n", board.flags, mines)
		s := bufio.NewScanner(os.Stdin)
		fmt.Println("Would you like to mark or dig?")
		s.Scan()
		answer := s.Text()
		switch strings.ToLower(answer) {
		case "mark":
			markSquare(board, s)
		case "dig":
			if !digSquare(board, s) {
				fmt.Println("You loose! Try again!")
				return
			}
		default:
			fmt.Println("Unknown response...")
		}
		time.Sleep(1 * time.Second)
		if board.CheckWin() {
			fmt.Println("Congratulations! You Win!")
			return
		}
	}
}

func markSquare(board *Board, s *bufio.Scanner) {
	fmt.Println("I should mark the board...")
	coords, err := getCoords(s)
	if err != nil {
		return
	}
	board.Mark(coords.x, coords.y)
}

func digSquare(board *Board, s *bufio.Scanner) bool {
	fmt.Println("I should dig the board...")
	coords, err := getCoords(s)
	if err != nil {
		return true
	}
	fmt.Println(coords)
	if !board.Dig(coords.x, coords.y) {
		fmt.Println("I hit a bomb...")
		return false
	}
	return true
}

func getCoords(s *bufio.Scanner) (*Point, error) {
	fmt.Println("Which x?")
	s.Scan()
	x, err := strconv.Atoi(s.Text())
	if err != nil {
		return nil, err
	}
	fmt.Println("Which y?")
	s.Scan()
	y, err := strconv.Atoi(s.Text())
	if err != nil {
		return nil, err
	}
	return &Point{
		x: x,
		y: y,
	}, nil
}

type Point struct {
	x int
	y int
}

func genArr(width, height, firstX, firstY int) []*Point {
	ret := make([]*Point, 0)
	for i := 0; i < height; i++ {
		for j := 0; j < width; j++ {
			if j == firstX && i == firstY {
				continue
			}
			ret = append(ret, &Point{
				x: j,
				y: i,
			})
		}
	}
	return ret
}
