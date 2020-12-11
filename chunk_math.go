package throwpro

import (
	"fmt"
	"log"
	"math"
	"math/rand"
)

var rings = [][2]int{{1408, 2688}, {4480, 5760}, {7552, 8832}, {10624, 11904}, {13696, 14976}, {16768, 18048}, {19840, 21120}, {22912, 24192}}

func ChunkFromCenter(x, y int) Chunk {
	return Chunk{(x - modLikePython(x, 16)) / 16, (y - modLikePython(y, 16)) / 16}
}

func ChunkFromPosition(x, y float64) Chunk {
	return Chunk{(int(x) - modLikePython(int(x), 16)) / 16, (int(y) - modLikePython(int(y), 16)) / 16}
}

func (c Chunk) Staircase() (int, int) {
	x, y := c.Center()
	return x - 4, y - 4
}

func (c Chunk) Distance(other interface{}) float64 {
	return c.ChunkDist(other.(Chunk))
}

func (c Chunk) GetID() string {
	return fmt.Sprintf(`%d,%d`, c[0], c[1])
}

func (c Chunk) String() string {
	x, y := c.Center()
	ring := RingID(c)
	return fmt.Sprintf("chunk %d,%d (center %d, %d, ring %d)", c[0], c[1], x, y, ring)
}

func RingID(c Chunk) int {
	cDist := c.Dist(0, 0)
	for n, ring := range rings {
		minDist, maxDist := float64(ring[0]), float64(ring[1])
		if cDist < minDist-240 {
			continue
		}
		if cDist > maxDist+240 {
			continue
		}
		return n
	}
	return -1
}

type Layer func([]Throw, Chunk) int

type LayerSet struct {
	AnglePref       float64
	RingMod         float64
	AverageDistance float64
	MathFactor      float64
}

func (ls LayerSet) Mutate() LayerSet {
	factor := 0.20
	eff := (rand.Float64() - .5) * 2 * factor
	switch rand.Intn(4) {
	case 0:
		ls.AnglePref *= 1 + eff
	case 1:
		ls.AverageDistance *= 1 + eff
	case 2:
		ls.RingMod *= 1 + eff
	case 3:
		ls.MathFactor *= 1 + eff
	}
	return ls
}

func (ls LayerSet) Layers() []Layer {
	return []Layer{ls.Angle, ls.Ring, ls.CrossAngle}
}

func (ls LayerSet) Ring(t []Throw, c Chunk) int {
	ringID := RingID(c)
	if ringID == -1 {
		return 0
	}
	cDist := c.Dist(0, 0)
	minDist, maxDist := float64(rings[ringID][0]), float64(rings[ringID][1])
	preferred := minDist + (maxDist-minDist)*ls.AverageDistance
	ring := cDist - preferred
	if ring < ls.RingMod {
		return 3
	}
	if ring < ls.RingMod*2 {
		return 2
	}
	return 1
}

func (ls LayerSet) Angle(ts []Throw, c Chunk) int {
	total := 0
	for _, t := range ts {
		delta := math.Abs(c.Angle(t.A, t.X, t.Y))
		if delta > radsFromDegs(.7) {
			return 0
		}
		if delta < ls.AnglePref {
			total += 4
		}
		if delta < ls.AnglePref*2 {
			total += 3
		}
		if delta < ls.AnglePref*3 {
			total += 2
		}
		total += 1
	}
	return total
}

func (ls LayerSet) CrossAngle(ts []Throw, c Chunk) int {
	if len(ts) <= 1 {
		return 1
	}
	printout := rand.Intn(10000) == 0
	if c == DEBUG_CHUNK {
		printout = true
	}
	if !DEBUG {
		printout = false
	}
	score := 1
	ax := 0.0
	ay := 0.0
	for n, t := range ts[:len(ts)-1] {
		for _, ot := range ts[n+1:] {
			k := ((ot.Y-t.Y)*math.Sin(ot.A) + (ot.X-t.X)*math.Cos(ot.A)) / math.Sin(ot.A-t.A)
			ny := t.Y + k*math.Cos(t.A)
			nx := t.X - k*math.Sin(t.A)

			distFromPerfect := c.Dist(nx, ny)

			if printout {
				log.Printf("chunk %s crossangle %.1f %.1f dist %.1f", c, ax, ay, distFromPerfect)
				log.Println("throws", ts)
				log.Println("debug chunk", DEBUG_CHUNK)
			}
			if distFromPerfect < ls.MathFactor {
				score += 4
			}
			if distFromPerfect < ls.MathFactor*5 {
				score += 3
			}
			if distFromPerfect < ls.MathFactor*12 {
				score += 2
			}
			if distFromPerfect < ls.MathFactor*25 {
				score += 1
			}

			ay += ny
			ax += nx
			if printout {
				log.Printf("chunk %s crossangle %.1f %.1f", c, nx, ny)
			}
		}
	}
	// ax /= float64(len(ts))
	// ay /= float64(len(ts))

	return score
}

func dist(x, y, x2, y2 float64) float64 {
	dx := x - x2
	dy := y - y2
	return math.Sqrt(dx*dx + dy*dy)
}

func (c Chunk) Dist(x, y float64) float64 {
	cx, cy := c.Center()
	return dist(float64(cx), float64(cy), x, y)
}

func (c Chunk) Angle(a, sx, sy float64) float64 {
	x, y := c.Center()
	atan := math.Atan2(sx-float64(x), float64(y)-sy) + math.Pi*2
	atan = math.Mod(atan, math.Pi*2)
	diff := wrapRads(a - atan)
	return diff
}

func (c Chunk) Center() (int, int) {
	return c[0]*16 + 8, c[1]*16 + 8
}

func radsFromDegs(degs float64) float64 {
	return wrapRads(degs * (math.Pi / 180))
}

func wrapRads(rads float64) float64 {
	for rads < math.Pi {
		rads += math.Pi * 2
	}
	for rads > math.Pi {
		rads -= math.Pi * 2
	}
	return rads
}

func ChunksInThrow(t Throw) ChunkList {
	angle := t.A
	cx, cy := t.X, t.Y
	dx, dy := -math.Sin(angle), math.Cos(angle)

	chunks := make(ChunkList, 0)
	chunksFound := map[Chunk]bool{}
	for {
		blockX := int(math.Floor(cx))
		blockY := int(math.Floor(cy))

		centerX := modLikePython(blockX, 16)
		centerY := modLikePython(blockY, 16)

		for xo := -1; xo < 1; xo++ {
			for yo := -1; yo < 1; yo++ {
				chunk := Chunk{(blockX-centerX)/16 + xo, (blockY-centerY)/16 + yo}
				if _, found := chunksFound[chunk]; !found {
					chunksFound[chunk] = true
					chunks = append(chunks, chunk)
				}
			}
		}

		lastDist := dist(0, 0, cx, cy)
		cx += dx * 2
		cy += dy * 2
		newDist := dist(0, 0, cx, cy)
		if newDist > lastDist && newDist > float64(rings[len(rings)-1][1]+240) {
			break
		}
	}
	return chunks
}

func modLikePython(d, m int) int {
	var res int = d % m
	if (res < 0 && m > 0) || (res > 0 && m < 0) {
		return res + m
	}
	return res
}
