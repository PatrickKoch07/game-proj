package colliders

import (
	"slices"
	"sync"

	"github.com/PatrickKoch07/game-proj/internal/gameState"
	"github.com/PatrickKoch07/game-proj/internal/logger"
)

type colliderMapCoords struct {
	X int
	Y int
}

type ColliderMap2D struct {
	Map    map[colliderMapCoords][]*Collider2D
	ChunkX int
	ChunkY int
}

// Objects must read in the whole state of all collidermaps and make changes to the whole state
type colliderMapLayers struct {
	FlagToMap map[gameState.Flag]*ColliderMap2D
	Mu        sync.Mutex
}

var colliderFlagToMap *colliderMapLayers
var onceFlagToMap sync.Once

func (cm *ColliderMap2D) worldCoordsToColliderCoords(xy WorldCoords) colliderMapCoords {
	return colliderMapCoords{X: int(xy.X) / cm.ChunkX, Y: int(xy.Y) / cm.ChunkY}
}

func (cm *ColliderMap2D) getColliderCoords(collider *Collider2D) []colliderMapCoords {
	topWorldY := collider.CenterCoords.Y + collider.Height/2.0
	bottomWorldY := collider.CenterCoords.Y - collider.Height/2.0
	leftWorldX := collider.CenterCoords.X - collider.Width/2.0
	rightWorldX := collider.CenterCoords.X + collider.Width/2.0

	topLeft := cm.worldCoordsToColliderCoords(WorldCoords{X: leftWorldX, Y: topWorldY})
	bottomLeft := cm.worldCoordsToColliderCoords(WorldCoords{X: leftWorldX, Y: bottomWorldY})
	topRight := cm.worldCoordsToColliderCoords(WorldCoords{X: rightWorldX, Y: topWorldY})

	mapCoordsColliderTouches := make(
		[]colliderMapCoords, (topRight.X-topLeft.X+1)*(topLeft.Y-bottomLeft.Y+1),
	)

	counter := 0
	for dx := 0; dx < (topRight.X - topLeft.X + 1); dx++ {
		for dy := 0; dy < (topLeft.Y - bottomLeft.Y + 1); dy++ {
			mapCoordsColliderTouches[counter] = colliderMapCoords{
				X: bottomLeft.X + dx,
				Y: bottomLeft.Y + dy,
			}
			counter++
		}
	}

	return mapCoordsColliderTouches
}

func (cm *ColliderMap2D) getPrevColliderCoords(
	collider *Collider2D, prevCenter WorldCoords,
) []colliderMapCoords {
	topWorldY := prevCenter.Y + collider.Height/2.0
	bottomWorldY := prevCenter.Y - collider.Height/2.0
	leftWorldX := prevCenter.X - collider.Width/2.0
	rightWorldX := prevCenter.X + collider.Width/2.0

	topLeft := cm.worldCoordsToColliderCoords(WorldCoords{X: leftWorldX, Y: topWorldY})
	bottomLeft := cm.worldCoordsToColliderCoords(WorldCoords{X: leftWorldX, Y: bottomWorldY})
	topRight := cm.worldCoordsToColliderCoords(WorldCoords{X: rightWorldX, Y: topWorldY})

	mapCoordsColliderTouches := make(
		[]colliderMapCoords, (topRight.X-topLeft.X+1)*(topLeft.Y-bottomLeft.Y+1),
	)

	counter := 0
	for dx := 0; dx < (topRight.X - topLeft.X + 1); dx++ {
		for dy := 0; dy < (topLeft.Y - bottomLeft.Y + 1); dy++ {
			mapCoordsColliderTouches[counter] = colliderMapCoords{
				X: bottomLeft.X + dx,
				Y: bottomLeft.Y + dy,
			}
			counter++
		}
	}

	return mapCoordsColliderTouches
}

// not safe
func updateColliderInMap(collider *Collider2D, prevWorldCoord WorldCoords) {
	for _, flag := range collider.Tags {
		colliderMap, ok := getColliderMap(flag)
		if !ok {
			continue
		}

		// remove
		colliderMapCoords := colliderMap.getPrevColliderCoords(collider, prevWorldCoord)
		for _, colliderMapCoord := range colliderMapCoords {
			colliderMap.Map[colliderMapCoord] = slices.DeleteFunc(
				colliderMap.Map[colliderMapCoord],
				func(mappedCollider *Collider2D) bool { return equals(collider, mappedCollider) },
			)
		}

		// add
		colliderMapCoords = colliderMap.getColliderCoords(collider)
		for _, colliderMapCoord := range colliderMapCoords {
			colliderMap.Map[colliderMapCoord] = append(colliderMap.Map[colliderMapCoord], collider)
		}
	}
	colliderMap, ok := getColliderMap(gameState.AllColliders)
	if !ok {
		logger.LOG.Error().Msg("Something wrong collider maps: no base allcollider map")
		return
	}
	colliderMapCoords := colliderMap.getColliderCoords(collider)

	// remove
	colliderMapCoords = colliderMap.getPrevColliderCoords(collider, prevWorldCoord)
	for _, colliderMapCoord := range colliderMapCoords {
		colliderMap.Map[colliderMapCoord] = slices.DeleteFunc(
			colliderMap.Map[colliderMapCoord],
			func(mappedCollider *Collider2D) bool { return equals(collider, mappedCollider) },
		)
	}

	// add
	for _, colliderMapCoord := range colliderMapCoords {
		colliderMap.Map[colliderMapCoord] = append(colliderMap.Map[colliderMapCoord], collider)
	}
}

// not safe
func removeColliderFromMaps(collider *Collider2D) {
	for _, flag := range collider.Tags {
		colliderMap, ok := getColliderMap(flag)
		if !ok {
			continue
		}
		colliderMapCoords := colliderMap.getColliderCoords(collider)
		for _, colliderMapCoord := range colliderMapCoords {
			colliderMap.Map[colliderMapCoord] = slices.DeleteFunc(
				colliderMap.Map[colliderMapCoord],
				func(mappedCollider *Collider2D) bool { return equals(collider, mappedCollider) },
			)
		}
	}
	colliderMap, ok := getColliderMap(gameState.AllColliders)
	if !ok {
		logger.LOG.Error().Msg("Something wrong collider maps: no base allcollider map")
		return
	}
	colliderMapCoords := colliderMap.getColliderCoords(collider)
	for _, colliderMapCoord := range colliderMapCoords {
		colliderMap.Map[colliderMapCoord] = slices.DeleteFunc(
			colliderMap.Map[colliderMapCoord],
			func(mappedCollider *Collider2D) bool { return equals(collider, mappedCollider) },
		)
	}
}

func getColliderMapLayers() *colliderMapLayers {
	if colliderFlagToMap == nil {
		onceFlagToMap.Do(createColliderFlagToMap)
	}
	return colliderFlagToMap
}

// not safe
func getColliderMap(flag gameState.Flag) (*ColliderMap2D, bool) {
	colliderMap, ok := getColliderMapLayers().FlagToMap[flag]
	return colliderMap, ok
}

func createColliderFlagToMap() {
	colliderFlagToMap = new(colliderMapLayers)
	colliderFlagToMap.FlagToMap = make(map[gameState.Flag]*ColliderMap2D)
	colliderMap := &ColliderMap2D{
		Map:    make(map[colliderMapCoords][]*Collider2D),
		ChunkX: 64,
		ChunkY: 64,
	}
	colliderFlagToMap.FlagToMap[gameState.AllColliders] = colliderMap

	// TODO technically this stuff should be the game side of things, rather than game engine since
	// this can and should be changed for each game. I guess I could be lazy and use fixed layers
	// like unreal

	colliderMap = &ColliderMap2D{
		Map:    make(map[colliderMapCoords][]*Collider2D),
		ChunkX: 64,
		ChunkY: 64,
	}
	colliderFlagToMap.FlagToMap[gameState.EnvironmentCollider] = colliderMap
	// etc.
}

// not safe
func AddColliderToMaps(collider *Collider2D) {
	for _, flag := range collider.Tags {
		colliderMap, ok := getColliderMap(flag)
		if !ok {
			continue
		}
		allColliderMapCoords := colliderMap.getColliderCoords(collider)
		for _, colliderMapCoord := range allColliderMapCoords {
			colliderMap.Map[colliderMapCoord] = append(colliderMap.Map[colliderMapCoord], collider)
		}
	}
	colliderMap, ok := getColliderMap(gameState.AllColliders)
	if !ok {
		logger.LOG.Error().Msg("Something wrong collider maps: no base allcollider map")
		return
	}
	allColliderMapCoords := colliderMap.getColliderCoords(collider)
	for _, colliderMapCoord := range allColliderMapCoords {
		colliderMap.Map[colliderMapCoord] = append(colliderMap.Map[colliderMapCoord], collider)
	}
}
