package defusedivision

type Cell struct {
	Contents  string           `json:"contents"`
	X         int              `json:"x"`
	Y         int              `json:"y"`
	Probed    bool             `json:"probed"`
	Flagged   bool             `json:"flagged"`
	Neighbors map[string]*bool `json:"neighbors"`
}

type Minefield struct {
	Height    int     `json:"height"`
	Width     int     `json:"width"`
	Minecount int     `json:"mine_count"`
	Selected  []int   `json:"selected"`
	Victory   bool    `json:"victory"`
	Cells     []*Cell `json:"cells"`
}

type Player struct {
	Name   string    `json:"name"`
	Living bool      `json:"living"`
	Field  Minefield `json:"minefield"`
}

type State struct {
	Ready   bool              `json:"ready"`
	Players map[string]Player `json:"players"`
}
