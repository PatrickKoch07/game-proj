package colliders

import (
	"math"
	"slices"
	"sync"

	"github.com/PatrickKoch07/game-proj/internal/gameState"
	"github.com/PatrickKoch07/game-proj/internal/logger"
	"github.com/PatrickKoch07/game-proj/internal/scenes"
	"github.com/PatrickKoch07/game-proj/internal/utils"
)

const spaceBetweenRays float32 = 1.0
const maxCollisionsPerMove int = 100

type WorldCoords struct {
	X float32
	Y float32
}

type Collider2D struct {
	Tags []gameState.Flag
	// where this collider lives
	CenterCoords WorldCoords
	Width        float32
	Height       float32
	// On collide functions.
	// Note: this is when the collider boundaries TOUCH, not when they enclose each other.
	// Ex. a small object enters a very large object:
	// On the way inside, the on enter and on exit will be called.
	// Inside, nothing is called.
	// On the way out, the on enter and on exit will be called again.
	OnEnterCollision func(*Collider2D)
	OnExitCollision  func(*Collider2D)
	// how to interact with other colliders
	// this in particular stops colliders from overlaping & only calls the OnEnterCollision
	Block  []gameState.Flag
	Ignore []gameState.Flag
	Parent *scenes.GameObject
}

func equals(c1 *Collider2D, c2 *Collider2D) bool {
	// case: same center coords and same dimensions:
	// This should only happen if the objects can overlap
	// If the objects have the same parent and same tags, then they should've been combined?
	// going with this logic for now (not sure how annoying it is to combine colliders)
	if c1.CenterCoords != c2.CenterCoords {
		return false
	}
	if c1.Width != c2.Width {
		return false
	}
	if c1.Height != c2.Height {
		return false
	}
	if slices.Compare(c1.Tags, c2.Tags) != 0 {
		return false
	}
	if *c1.Parent != *c2.Parent {
		return false
	}
	return true
}

// Moves the collider until it encounters a blocking collider. Along the way, it notifies all
// colliders it encounters, excluding colliders with any of the ignored tags. If it hits any
// blocking object along its movement path, all movement stops.
func (c *Collider2D) MoveCollider(
	finalCenter WorldCoords,
) WorldCoords {
	deltaY := finalCenter.Y - c.CenterCoords.Y
	deltaX := finalCenter.X - c.CenterCoords.X
	// split the big move into distinct points to test for collisions
	var dx, dy float32
	if deltaX == 0 {
		if deltaY == 0 {
			return c.CenterCoords
		}
		dx = 0.0
		dy = spaceBetweenRays
	} else {
		slope := math.Abs(float64(deltaY / deltaX))
		dx = float32(math.Sqrt(float64(spaceBetweenRays*spaceBetweenRays) / (1.0 + slope*slope)))
		dy = dx * float32(slope)
	}

	// Locking ALL collider maps so we can safely check for collisions without one randomly
	// popping in (or one randomly popping into half only, so we lock all)
	// Theres gotta be another way, but oh well
	getColliderMapLayers().Mu.Lock()
	initialCenter := WorldCoords{X: c.CenterCoords.X, Y: c.CenterCoords.Y}
	defer func() {
		updateColliderInMap(c, initialCenter)
		getColliderMapLayers().Mu.Unlock()
	}()

	// get all collider maps to check blocking collisions against
	colliderMaps := make([]*ColliderMap2D, len(c.Block)+1)
	for i, blockFlag := range c.Block {
		colliderMap, ok := getColliderMap(blockFlag)
		if !ok {
			logger.LOG.Error().Msg("Bad collider map flag in collision detection.")
			continue
		}
		colliderMaps[i] = colliderMap
	}

	// add the total collider map for overlaps and check the current position for any overlaps
	colliderMap, ok := getColliderMap(gameState.AllColliders)
	colliderMaps[len(colliderMaps)-1] = colliderMap
	if !ok {
		logger.LOG.Error().Msg("Bad collider map flag in collision detection, can't call overlaps")
	}
	var lastSeenColliders []*Collider2D = make([]*Collider2D, 0, maxCollisionsPerMove)
	var wg sync.WaitGroup
	collidedCh := make(chan *Collider2D, maxCollisionsPerMove)
	upLeft := WorldCoords{X: c.CenterCoords.X - c.Width/2.0, Y: c.CenterCoords.Y + c.Height/2.0}
	upRight := WorldCoords{X: c.CenterCoords.X + c.Width/2.0, Y: c.CenterCoords.Y + c.Height/2.0}
	downLeft := WorldCoords{X: c.CenterCoords.X - c.Width/2.0, Y: c.CenterCoords.Y - c.Height/2.0}
	downRight := WorldCoords{X: c.CenterCoords.X + c.Width/2.0, Y: c.CenterCoords.Y - c.Height/2.0}
	mapIds := colliderMap.getColliderCoords(c)
	for _, id := range mapIds {
		for _, collider := range colliderMap.Map[id] {
			if equals(c, collider) {
				continue
			}
			if utils.AnyOverlap(c.Ignore, collider.Tags) {
				continue
			}
			wg.Add(1)
			go func() {
				defer wg.Done()
				verticalEdgeIntersec(upLeft, upRight, collider, collidedCh)
				verticalEdgeIntersec(downLeft, downRight, collider, collidedCh)
				horizontalEdgeIntersec(downLeft, upLeft, collider, collidedCh)
				horizontalEdgeIntersec(downRight, upRight, collider, collidedCh)
			}()
		}
	}
	wg.Wait()
	close(collidedCh)
	// Loop over overlapping objects and add to objects seen
	if len(collidedCh) == maxCollisionsPerMove {
		logger.LOG.Warn().Msgf("Saw way too many collisions moving collider: %v", c)
	}
	for collider := range collidedCh {
		if slices.ContainsFunc(lastSeenColliders, func(nc *Collider2D) bool { return equals(nc, collider) }) {
			continue
		}
		// logger.LOG.Debug().Msg("We started overlapped with one collider")
		lastSeenColliders = append(lastSeenColliders, collider)
	}

	// iterates over the discretized movement points
	currentCenter := WorldCoords{X: c.CenterCoords.X, Y: c.CenterCoords.Y}
	for i := 0; currentCenter.Y != finalCenter.Y || currentCenter.X != finalCenter.X; i++ {
		// returned if this step hits a blocking collision
		previousCenter := WorldCoords{X: currentCenter.X, Y: currentCenter.Y}

		dx = min(dx, float32(math.Abs(float64(finalCenter.X-currentCenter.X))))
		dy = min(dy, float32(math.Abs(float64(finalCenter.Y-currentCenter.Y))))

		if deltaX > 0 {
			currentCenter.X += dx
		} else {
			currentCenter.X -= dx
		}
		if deltaY > 0 {
			currentCenter.Y += dy
		} else {
			currentCenter.Y -= dy
		}

		// we will only be checking the collisions on the sides leading the movement.
		// Ex. if we move up and left, only check for collisions on the upper-most and left sides
		var verticalDown WorldCoords
		var verticalUp WorldCoords
		if deltaX > 0 {
			verticalDown = WorldCoords{X: currentCenter.X + c.Width/2.0, Y: currentCenter.Y - c.Height/2.0}
			verticalUp = WorldCoords{X: currentCenter.X + c.Width/2.0, Y: currentCenter.Y + c.Height/2.0}
		} else {
			verticalDown = WorldCoords{X: currentCenter.X - c.Width/2.0, Y: currentCenter.Y - c.Height/2.0}
			verticalUp = WorldCoords{X: currentCenter.X - c.Width/2.0, Y: currentCenter.Y + c.Height/2.0}
		}
		var horizontalLeft WorldCoords
		var horizontalRight WorldCoords
		if deltaY > 0 {
			horizontalLeft = WorldCoords{X: currentCenter.X - c.Width/2.0, Y: currentCenter.Y + c.Height/2.0}
			horizontalRight = WorldCoords{X: currentCenter.X + c.Width/2.0, Y: currentCenter.Y + c.Height/2.0}
		} else {
			horizontalLeft = WorldCoords{X: currentCenter.X - c.Width/2.0, Y: currentCenter.Y - c.Height/2.0}
			horizontalRight = WorldCoords{X: currentCenter.X + c.Width/2.0, Y: currentCenter.Y - c.Height/2.0}
		}

		// All go routines here write to buffered channel IF they have a collision.
		// They write their own collider so we can loop through all the colliders we 'hit'
		// If we have more than 'max' writes to objects we collided with, there is an issue (maybe)
		collidedCh = make(chan *Collider2D, maxCollisionsPerMove)
		for _, colliderMap := range colliderMaps[:len(colliderMaps)-1] {
			horizontalIds := horizontalEdgeIds(horizontalLeft, horizontalRight, colliderMap)
			for _, horizontalId := range horizontalIds {
				for _, collider := range colliderMap.Map[horizontalId] {
					if equals(c, collider) {
						continue
					}
					wg.Add(1)
					go func() {
						defer wg.Done()
						verticalEdgeIntersec(horizontalLeft, horizontalRight, collider, collidedCh)
					}()
				}
			}
			verticalIds := verticalEdgeIds(verticalDown, verticalUp, colliderMap)
			for _, verticalId := range verticalIds {
				for _, collider := range colliderMap.Map[verticalId] {
					if equals(c, collider) {
						continue
					}
					wg.Add(1)
					go func() {
						defer wg.Done()
						horizontalEdgeIntersec(verticalDown, verticalUp, collider, collidedCh)
					}()
				}
			}
		}
		wg.Wait()
		close(collidedCh)

		if len(collidedCh) == maxCollisionsPerMove {
			logger.LOG.Warn().Msgf("Saw way too many collisions moving collider: %v", c)
		}
		// this is to prevent double calling for a single object (both edges touch somehow)
		blockColliders := make([]*Collider2D, 0, maxCollisionsPerMove)
		for collider := range collidedCh {
			if slices.ContainsFunc(blockColliders, func(bc *Collider2D) bool { return equals(bc, collider) }) {
				continue
			}
			blockColliders = append(blockColliders, collider)
			go c.OnEnterCollision(collider)
			go collider.OnEnterCollision(c)
		}
		if len(blockColliders) != 0 {
			logger.LOG.Info().Msg("Blocked collider movement.")
			c.CenterCoords = previousCenter
			return previousCenter
		}

		// If no blocking collisions, check the colliderMap for notifying collisions
		colliderMap = colliderMaps[len(colliderMaps)-1]
		collidedCh = make(chan *Collider2D, maxCollisionsPerMove)
		horizontalIds := horizontalEdgeIds(horizontalLeft, horizontalRight, colliderMap)
		for _, horizontalId := range horizontalIds {
			for _, collider := range colliderMap.Map[horizontalId] {
				if equals(c, collider) {
					continue
				}
				if utils.AnyOverlap(c.Ignore, collider.Tags) {
					continue
				}
				wg.Add(1)
				go func() {
					defer wg.Done()
					verticalEdgeIntersec(horizontalLeft, horizontalRight, collider, collidedCh)
				}()
			}
		}
		verticalIds := verticalEdgeIds(verticalDown, verticalUp, colliderMap)
		for _, verticalId := range verticalIds {
			for _, collider := range colliderMap.Map[verticalId] {
				if equals(c, collider) {
					continue
				}
				if utils.AnyOverlap(c.Ignore, collider.Tags) {
					continue
				}
				wg.Add(1)
				go func() {
					defer wg.Done()
					horizontalEdgeIntersec(verticalDown, verticalUp, collider, collidedCh)
				}()
			}
		}
		wg.Wait()
		close(collidedCh)

		// Loop over hit objects and add to objects seen this iteration (no duplicates)
		seenColliders := make([]*Collider2D, 0, maxCollisionsPerMove)
		if len(collidedCh) == maxCollisionsPerMove {
			logger.LOG.Warn().Msgf("Saw way too many collisions moving collider: %v", c)
		}
		for collider := range collidedCh {
			if slices.ContainsFunc(seenColliders, func(bc *Collider2D) bool { return equals(bc, collider) }) {
				continue
			}
			seenColliders = append(seenColliders, collider)
		}

		// for each object seen last position, check if we didn't see it hit again (call exit!)
		var lastSeenCollidersBuffer []*Collider2D
		for _, collider := range lastSeenColliders {
			hadCollision := false
			for i, seenCollider := range seenColliders {
				if !equals(collider, seenCollider) {
					continue
				}
				hadCollision = true
				// pop element just seen; only want newly seen elements by the end of this block
				seenColliders[i] = seenColliders[len(seenColliders)-1]
				seenColliders = seenColliders[:len(seenColliders)-1]
			}
			// if no collision then we left
			if !hadCollision {
				go c.OnExitCollision(collider)
				go collider.OnExitCollision(collider)
				// if it did, then we leave it as something seen
			} else {
				lastSeenCollidersBuffer = append(lastSeenCollidersBuffer, collider)
			}
		}

		// Each object is seen for the first time b/c they werent popped
		for _, collider := range seenColliders {
			go c.OnEnterCollision(collider)
			go collider.OnEnterCollision(c)
			lastSeenCollidersBuffer = append(lastSeenCollidersBuffer, collider)
		}
		lastSeenColliders = lastSeenCollidersBuffer
	}
	c.CenterCoords = finalCenter

	return finalCenter
}

// TODO:
// // Moves the collider until it encounters a blocking collider. Along the way, it notifies all
// // colliders it encounters, excluding colliders with any of the ignored tags. In addition, this
// // continues to move the collider along any non-blocking axis.
// // Ex. A character wishes to move diagonally UP and LEFT, but is blocked by a vertical wall. The
// // character will continue to slide UP the wall, but not progress anymore LEFT
// func (c *Collider2D) MoveColliderWithSlide() {}

// TODO:
// function for finding when are inside a collider/changing current implementation to exit only
// when no longer touching the volume (and not inside of it)

func horizontalEdgeIds(left, right WorldCoords, colliderMap *ColliderMap2D) []colliderMapCoords {
	LeftCoords := colliderMap.worldCoordsToColliderCoords(left)
	RightCoords := colliderMap.worldCoordsToColliderCoords(right)
	length := RightCoords.X - LeftCoords.X + 1
	ColliderMapCoords := make([]colliderMapCoords, length)
	for i := 0; i < length; i++ {
		ColliderMapCoords[i] = colliderMapCoords{X: LeftCoords.X + i, Y: LeftCoords.Y}
	}
	return ColliderMapCoords
}

func verticalEdgeIds(down, up WorldCoords, colliderMap *ColliderMap2D) []colliderMapCoords {
	downCoords := colliderMap.worldCoordsToColliderCoords(down)
	upCoords := colliderMap.worldCoordsToColliderCoords(up)
	length := upCoords.Y - downCoords.Y + 1
	ColliderMapCoords := make([]colliderMapCoords, length)
	for i := 0; i < length; i++ {
		ColliderMapCoords[i] = colliderMapCoords{X: downCoords.X, Y: downCoords.Y + i}
	}
	return ColliderMapCoords
}

func horizontalEdgeIntersec(
	pDown, pUp WorldCoords, collider *Collider2D, collidedCh chan<- *Collider2D,
) {
	downY := collider.CenterCoords.Y - collider.Height/2.0
	upY := collider.CenterCoords.Y + collider.Height/2.0
	// impossible to collide; removes the N regions
	/*
		Up  collider  down
		Y				N
		Y    ||||| 		Y
		N				Y
	*/
	if (pUp.Y < downY) || (pDown.Y > upY) {
		return
	}
	// if x-axis cant intersect; pup.x == pdown.x
	if (pUp.X < collider.CenterCoords.X-collider.Width/2.0) ||
		(pUp.X > collider.CenterCoords.X+collider.Width/2.0) {
		return
	}
	// collision if U is in the upmost or if D is in the downmost in the above example
	if (pDown.Y <= downY) || (pUp.Y >= upY) {
		collidedCh <- collider
		return
	}

	return
}

func verticalEdgeIntersec(
	pLeft, pRight WorldCoords, collider *Collider2D, collidedCh chan<- *Collider2D,
) {
	leftX := collider.CenterCoords.X - collider.Width/2.0
	rightX := collider.CenterCoords.X + collider.Width/2.0
	// impossible to collide; removes the N regions
	/*
		L:			Y | Y | N
		collider:  	  |||||
		R:			N | Y | Y
	*/
	if (pRight.X < leftX) || (pLeft.X > rightX) {
		return
	}
	// if y-axis cant intersect; pleft.y == pright.y
	if (pLeft.Y < collider.CenterCoords.Y-collider.Height/2.0) ||
		(pLeft.Y > collider.CenterCoords.Y+collider.Height/2.0) {
		return
	}
	// collision if L is in the leftmost or if R is in the rightmost in the above example
	if (pLeft.X <= leftX) || (pRight.X >= rightX) {
		collidedCh <- collider
		return
	}

	return
}
