package main

import (
	"flag"
	"fmt"
	"strconv"
	"time"

	"github.com/nsf/termbox-go"
)

/*
 * The code is as follows:
 *
 * 111, 110, 101, 100, 011, 010, 001, 000
 *
 * E.g., rule 250 represents:
 *
 * 111 110 101 100 011 010 001 000
 *  1   1   1   1   1   0   1   0
 *
 * Since 0b11111010 is 250
 */

func Btoi(b bool) int {
	if b {
		return 1
	} else {
		return 0
	}
}

func neighbours(state []bool, index int) int {
	sum := 0
	var left, right int
	if index == 0 {
		left = Btoi(state[len(state)-1])
	} else {
		left = Btoi(state[index-1])
	}

	if index == len(state)-1 {
		right = 0
	} else {
		right = Btoi(state[index+1])
	}

	sum += left << 2
	sum += Btoi(state[index]) << 1
	sum += right

	return sum
}

func calcCellState(neighbourState, rule int) bool {
	ns := uint(neighbourState)
	mask := 1 << ns
	next := (rule & mask) >> ns
	if next == 1 {
		return true
	} else if next == 0 {
		return false
	}
	panic(string(next) + " is not valid cell state")
}

// Generate the next state from current state
func nextState(state []bool, rule int) []bool {
	next := make([]bool, len(state))

	for i := range state {
		neighbourState := neighbours(state, i)
		next[i] = calcCellState(neighbourState, rule)
	}

	return next
}

// When we first start, we want to generate the entire board, and then
// after that we can start animating. This function does the initial generation.
func initialGen(state [][]bool, rule int) {
	for i := range state {
		if i == 0 { // skip first row, as this is the initial condition
			continue
		}

		state[i] = nextState(state[i-1], rule)
	}
}

func draw(state [][]bool) {
	termbox.Clear(termbox.ColorBlack, termbox.ColorBlack)

	for i := range state {
		for j := range state[i] {
			var color termbox.Attribute
			if state[i][j] {
				color = termbox.ColorWhite
			} else {
				color = termbox.ColorBlack
			}
			termbox.SetCell(j, i, ' ', termbox.ColorWhite, color)
		}
	}

	termbox.Flush()
}

func update(state [][]bool, rule int) [][]bool {
	next := state[1:]
	next = append(next, nextState(next[len(next)-1], rule))
	return next
}

func main() {
	flag.Parse()

	rule, err := strconv.Atoi(flag.Arg(0))
	if err != nil {
		fmt.Println("Please supply a rule.")
		return
	}

	if rule < 0 || rule > 255 {
		fmt.Println(rule, "is not within the allowed range, please enter a number between 0-255.")
		return
	}

	err = termbox.Init()
	if err != nil {
		panic(err)
	}
	defer termbox.Close()

	w, h := termbox.Size()
	board := make([][]bool, h)
	for i := range board {
		board[i] = make([]bool, w)
		for j := range board[i] {
			board[i][j] = false
		}
	}

	// Initial condition is middle element of top row is on, all others are off
	board[0][w/2] = true
	initialGen(board, rule)

	eventQueue := make(chan termbox.Event)
	go func(evq chan termbox.Event) {
		for {
			evq <- termbox.PollEvent()
		}
	}(eventQueue)

	const framePeriod = time.Millisecond * 100
	lastUpdate := time.Now()

mainLoop:
	for {
		if time.Since(lastUpdate) < framePeriod { // rate limit updates
			continue
		}
		select {
		case ev := <-eventQueue:
			switch ev.Type {
			case 0:
				if ev.Ch == 'q' {
					break mainLoop
				}
			default:
				// Handle key somehow
			}
		default:
		}

		draw(board)
		board = update(board, rule)
		lastUpdate = time.Now()
	}
}
