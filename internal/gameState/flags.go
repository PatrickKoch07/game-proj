package gameState

type Flag int

const (
	// reserve 0 for the OFF state incase you use flags to point to other flags
	_ = iota; 
	CloseRequested Flag = iota

	NextScene  Flag = iota
	TitleScene Flag = iota
	WorldScene Flag = iota
)
