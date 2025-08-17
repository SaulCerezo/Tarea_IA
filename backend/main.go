package main

import (
	"container/heap"
	"encoding/json"
	"log"
	"math/rand"
	"net/http"
	"strconv"
	"strings"
	"time"
)

// ----- Puzzle utilities -----

type State [9]int

var goal State = State{1, 2, 3, 4, 5, 6, 7, 8, 0}

func stateKey(s State) string {
	var b strings.Builder
	for i, v := range s {
		if i > 0 {
			b.WriteByte(',')
		}
		b.WriteString(strconv.Itoa(v))
	}
	return b.String()
}

func manhattan(s State) int {
	sum := 0
	for i, v := range s {
		if v == 0 {
			continue
		}
		goalRow := (v - 1) / 3
		goalCol := (v - 1) % 3
		row := i / 3
		col := i % 3
		d := abs(goalRow-row) + abs(goalCol-col)
		sum += d
	}
	return sum
}

func abs(x int) int {
	if x < 0 {
		return -x
	}
	return x
}

type Neighbor struct {
	state State
	move  string
}

func neighbors(s State) []Neighbor {
	// Find blank (0)
	blank := 0
	for i, v := range s {
		if v == 0 {
			blank = i
			break
		}
	}
	r, c := blank/3, blank%3
	var out []Neighbor
	swap := func(i, j int) State {
		ns := s
		ns[i], ns[j] = ns[j], ns[i]
		return ns
	}
	if r > 0 { // up
		out = append(out, Neighbor{swap(blank, blank-3), "UP"})
	}
	if r < 2 { // down
		out = append(out, Neighbor{swap(blank, blank+3), "DOWN"})
	}
	if c > 0 { // left
		out = append(out, Neighbor{swap(blank, blank-1), "LEFT"})
	}
	if c < 2 { // right
		out = append(out, Neighbor{swap(blank, blank+1), "RIGHT"})
	}
	return out
}

// ----- A* implementation -----

type Node struct {
	state  State
	g, h   int
	index  int // index in the heap
	parent *Node
	move   string
}

// Priority Queue (min-heap by f = g + h)
type PriorityQueue []*Node

func (pq PriorityQueue) Len() int { return len(pq) }
func (pq PriorityQueue) Less(i, j int) bool {
	fi := pq[i].g + pq[i].h
	fj := pq[j].g + pq[j].h
	if fi == fj {
		// Tie-breaker: lower h preferred
		return pq[i].h < pq[j].h
	}
	return fi < fj
}
func (pq PriorityQueue) Swap(i, j int) {
	pq[i], pq[j] = pq[j], pq[i]
	pq[i].index = i
	pq[j].index = j
}
func (pq *PriorityQueue) Push(x interface{}) {
	n := len(*pq)
	item := x.(*Node)
	item.index = n
	*pq = append(*pq, item)
}
func (pq *PriorityQueue) Pop() interface{} {
	old := *pq
	n := len(old)
	item := old[n-1]
	item.index = -1
	*pq = old[:n-1]
	return item
}

func aStar(start State) ([][]int, []string, int, bool) {
	open := &PriorityQueue{}
	heap.Init(open)

	startNode := &Node{state: start, g: 0, h: manhattan(start), parent: nil, move: ""}
	heap.Push(open, startNode)

	closed := make(map[string]bool)
	gScore := make(map[string]int)
	gScore[stateKey(start)] = 0

	expanded := 0

	for open.Len() > 0 {
		current := heap.Pop(open).(*Node)
		expanded++

		if current.state == goal {
			// reconstruct path
			var path [][]int
			var actions []string
			for n := current; n != nil; n = n.parent {
				s := make([]int, 9)
				for i := 0; i < 9; i++ {
					s[i] = n.state[i]
				}
				path = append(path, s)
				if n.parent != nil {
					actions = append(actions, n.move)
				}
			}
			// reverse
			for i, j := 0, len(path)-1; i < j; i, j = i+1, j-1 {
				path[i], path[j] = path[j], path[i]
			}
			for i, j := 0, len(actions)-1; i < j; i, j = i+1, j-1 {
				actions[i], actions[j] = actions[j], actions[i]
			}
			return path, actions, expanded, true
		}

		ckey := stateKey(current.state)
		closed[ckey] = true

		for _, nb := range neighbors(current.state) {
			if closed[stateKey(nb.state)] {
				continue
			}
			tentativeG := current.g + 1
			nbKey := stateKey(nb.state)
			prevG, seen := gScore[nbKey]
			if !seen || tentativeG < prevG {
				node := &Node{
					state:  nb.state,
					g:      tentativeG,
					h:      manhattan(nb.state),
					parent: current,
					move:   nb.move,
				}
				gScore[nbKey] = tentativeG
				heap.Push(open, node)
			}
		}
	}
	return nil, nil, expanded, false
}

// Shuffle by doing N random legal moves from the goal state.
func shuffle(n int) State {
	rand.Seed(time.Now().UnixNano())
	s := goal
	lastMove := ""
	for i := 0; i < n; i++ {
		neigh := neighbors(s)
		// avoid undoing the last move if possible
		candidates := neigh[:0]
		for _, nb := range neigh {
			if (lastMove == "UP" && nb.move == "DOWN") ||
				(lastMove == "DOWN" && nb.move == "UP") ||
				(lastMove == "LEFT" && nb.move == "RIGHT") ||
				(lastMove == "RIGHT" && nb.move == "LEFT") {
				continue
			}
			candidates = append(candidates, nb)
		}
		if len(candidates) == 0 {
			candidates = neigh
		}
		choice := candidates[rand.Intn(len(candidates))]
		s = choice.state
		lastMove = choice.move
	}
	return s
}

// ----- HTTP API -----

type SolveRequest struct {
	Start []int `json:"start"`
}

type SolveResponse struct {
	Moves    [][]int  `json:"moves"`
	Actions  []string `json:"actions"`
	Cost     int      `json:"cost"`
	Expanded int      `json:"expanded"`
	Ok       bool     `json:"ok"`
	Message  string   `json:"message,omitempty"`
}

type ShuffleRequest struct {
	Steps int `json:"steps"`
}

type StateResponse struct {
	State []int `json:"state"`
}

func withCORS(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
		w.Header().Set("Access-Control-Allow-Methods", "GET,POST,OPTIONS")
		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusOK)
			return
		}
		next.ServeHTTP(w, r)
	})
}

func handleInit(w http.ResponseWriter, r *http.Request) {
	resp := StateResponse{State: []int{1, 2, 3, 4, 5, 6, 7, 8, 0}}
	writeJSON(w, http.StatusOK, resp)
}

func handleShuffle(w http.ResponseWriter, r *http.Request) {
	var req ShuffleRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid JSON"})
		return
	}
	if req.Steps <= 0 {
		req.Steps = 20
	}
	s := shuffle(req.Steps)
	resp := StateResponse{State: make([]int, 9)}
	for i := 0; i < 9; i++ {
		resp.State[i] = s[i]
	}
	writeJSON(w, http.StatusOK, resp)
}

func handleSolve(w http.ResponseWriter, r *http.Request) {
	var req SolveRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid JSON"})
		return
	}
	if len(req.Start) != 9 {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "start must be length 9"})
		return
	}
	var st State
	for i := 0; i < 9; i++ {
		st[i] = req.Start[i]
	}
	path, actions, expanded, ok := aStar(st)
	resp := SolveResponse{
		Moves:    path,
		Actions:  actions,
		Cost:     len(actions),
		Expanded: expanded,
		Ok:       ok,
	}
	if !ok {
		resp.Message = "no solution found (should not happen for 8-puzzle reached via legal moves)"
	}
	writeJSON(w, http.StatusOK, resp)
}

func writeJSON(w http.ResponseWriter, code int, v interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	enc := json.NewEncoder(w)
	enc.SetIndent("", "  ")
	_ = enc.Encode(v)
}

func main() {
	mux := http.NewServeMux()
	mux.HandleFunc("/api/init", handleInit)
	mux.HandleFunc("/api/shuffle", handleShuffle)
	mux.HandleFunc("/api/solve", handleSolve)

	addr := ":8080"
	log.Printf("8-puzzle Go API listening on %s\n", addr)
	if err := http.ListenAndServe(addr, withCORS(mux)); err != nil {
		log.Fatal(err)
	}
}
