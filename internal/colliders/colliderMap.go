package colliders

import (
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

var colliderFlagToMap *map[gameState.Flag]*ColliderMap2D
var onceFlagToMap sync.Once

func (cm ColliderMap2D) worldCoordsToColliderCoords(xy WorldCoords) colliderMapCoords {
	return colliderMapCoords{X: int(xy.X) / cm.ChunkX, Y: int(xy.Y) / cm.ChunkY}
}

func (cm ColliderMap2D) colliderToAllColliderCoords(collider *Collider2D) []colliderMapCoords {
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

func GetColliderMap(flag gameState.Flag) (*ColliderMap2D, bool) {
	if colliderFlagToMap == nil {
		onceFlagToMap.Do(createColliderFlagToMap)
	}
	colliderMap, ok := (*colliderFlagToMap)[flag]
	return colliderMap, ok
}

func createColliderFlagToMap() {
	colliderFlagToMapValue := make(map[gameState.Flag]*ColliderMap2D)
	colliderFlagToMap = &colliderFlagToMapValue
	colliderMap := &ColliderMap2D{
		Map:    make(map[colliderMapCoords][]*Collider2D),
		ChunkX: 64,
		ChunkY: 64,
	}
	colliderFlagToMapValue[gameState.AllColliders] = colliderMap

	// TODO technically this stuff should be the game side of things, rather than game engine since
	// this can and should be changed for each game. I guess I could be lazy and use fixed layers
	// like unreal

	colliderMap = &ColliderMap2D{
		Map:    make(map[colliderMapCoords][]*Collider2D),
		ChunkX: 64,
		ChunkY: 64,
	}
	colliderFlagToMapValue[gameState.EnvironmentCollider] = colliderMap
	// etc.
}

func AddColliderToMaps(collider *Collider2D) {
	for _, flag := range collider.Tags {
		colliderMap, ok := GetColliderMap(flag)
		if !ok {
			continue
		}
		allColliderMapCoords := colliderMap.colliderToAllColliderCoords(collider)
		for _, colliderMapCoord := range allColliderMapCoords {
			colliderMap.Map[colliderMapCoord] = append(colliderMap.Map[colliderMapCoord], collider)
		}
	}
	colliderMap, ok := GetColliderMap(gameState.AllColliders)
	if !ok {
		logger.LOG.Error().Msg("Something wrong collider maps: no base allcollider map")
		return
	}
	allColliderMapCoords := colliderMap.colliderToAllColliderCoords(collider)
	for _, colliderMapCoord := range allColliderMapCoords {
		colliderMap.Map[colliderMapCoord] = append(colliderMap.Map[colliderMapCoord], collider)
	}
}
