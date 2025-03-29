package gameState

type Flag int32

const (
	// reserve 0 for the OFF state incase you use flags to point to other flags
	_                   = iota
	CloseRequested Flag = iota

	NextScene    Flag = iota
	LoadingScene Flag = iota
	TitleScene   Flag = iota
	WorldScene   Flag = iota

	EnvironmentCollider Flag = iota
	AllColliders        Flag = iota
)
