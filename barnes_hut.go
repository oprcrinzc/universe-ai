package main

import (
	"math"
	"runtime"
	"sync"

	rl "github.com/gen2brain/raylib-go/raylib"
)

// OctreeNode represents a cubic node in the 3D Barnes-Hut spatial partitioning tree
type OctreeNode struct {
	Center       rl.Vector3
	HalfSize     float32
	TotalMass    float64
	CenterOfMass rl.Vector3
	Body         *Body
	Children     [8]*OctreeNode
	IsLeaf       bool
	Count        int
}

// NodePool provides pre-allocated octree nodes to eliminate GC heap allocation overhead
type NodePool struct {
	nodes []OctreeNode
	idx   int
}

var globalNodePool = &NodePool{
	nodes: make([]OctreeNode, 131072), // Pre-allocated 131k nodes (sufficient for 20k+ bodies)
}

func (p *NodePool) Reset() {
	p.idx = 0
}

func (p *NodePool) Alloc(center rl.Vector3, halfSize float32) *OctreeNode {
	if p.idx >= len(p.nodes) {
		more := make([]OctreeNode, len(p.nodes)*2)
		copy(more, p.nodes)
		p.nodes = more
	}
	n := &p.nodes[p.idx]
	p.idx++

	n.Center = center
	n.HalfSize = halfSize
	n.TotalMass = 0
	n.CenterOfMass = rl.NewVector3(0, 0, 0)
	n.Body = nil
	n.Children = [8]*OctreeNode{}
	n.IsLeaf = true
	n.Count = 0
	return n
}

// NewOctreeNode creates an empty octree node with given spatial boundaries
func NewOctreeNode(center rl.Vector3, halfSize float32) *OctreeNode {
	return &OctreeNode{
		Center:   center,
		HalfSize: halfSize,
		IsLeaf:   true,
		Count:    0,
	}
}

// Subdivide partitions this node into 8 cubic child octants
func (node *OctreeNode) Subdivide() {
	node.SubdivideWithPool(nil)
}

// SubdivideWithPool partitions this node using the node pool if available
func (node *OctreeNode) SubdivideWithPool(pool *NodePool) {
	quarter := node.HalfSize * 0.5
	for i := 0; i < 8; i++ {
		cx := node.Center.X
		cy := node.Center.Y
		cz := node.Center.Z

		if (i & 1) != 0 {
			cx += quarter
		} else {
			cx -= quarter
		}

		if (i & 2) != 0 {
			cy += quarter
		} else {
			cy -= quarter
		}

		if (i & 4) != 0 {
			cz += quarter
		} else {
			cz -= quarter
		}

		pos := rl.NewVector3(cx, cy, cz)
		if pool != nil {
			node.Children[i] = pool.Alloc(pos, quarter)
		} else {
			node.Children[i] = NewOctreeNode(pos, quarter)
		}
	}
}

// getOctantIndex returns the child octant index (0-7) for a 3D position
func (node *OctreeNode) getOctantIndex(pos rl.Vector3) int {
	idx := 0
	if pos.X >= node.Center.X {
		idx |= 1
	}
	if pos.Y >= node.Center.Y {
		idx |= 2
	}
	if pos.Z >= node.Center.Z {
		idx |= 4
	}
	return idx
}

// Insert adds a body into the octree recursively
func (node *OctreeNode) Insert(body *Body, depth int) {
	node.insertWithPool(body, depth, nil)
}

func (node *OctreeNode) insertWithPool(body *Body, depth int, pool *NodePool) {
	if body == nil || body.Mass <= 0 {
		return
	}

	// 1. Empty leaf node: store body directly
	if node.Count == 0 {
		node.Body = body
		node.TotalMass = body.Mass
		node.CenterOfMass = body.Position
		node.IsLeaf = true
		node.Count = 1
		return
	}

	// 2. Already occupied leaf: subdivide into children and push both bodies down
	if node.IsLeaf {
		dx := node.Body.Position.X - body.Position.X
		dy := node.Body.Position.Y - body.Position.Y
		dz := node.Body.Position.Z - body.Position.Z
		distSq := dx*dx + dy*dy + dz*dz

		if depth >= 22 || distSq < 0.0001 {
			newMass := node.TotalMass + body.Mass
			if newMass > 0 {
				invM := 1.0 / newMass
				node.CenterOfMass.X = float32((float64(node.CenterOfMass.X)*node.TotalMass + float64(body.Position.X)*body.Mass) * invM)
				node.CenterOfMass.Y = float32((float64(node.CenterOfMass.Y)*node.TotalMass + float64(body.Position.Y)*body.Mass) * invM)
				node.CenterOfMass.Z = float32((float64(node.CenterOfMass.Z)*node.TotalMass + float64(body.Position.Z)*body.Mass) * invM)
			}
			node.TotalMass = newMass
			node.Count++
			return
		}

		existingBody := node.Body
		node.Body = nil
		node.IsLeaf = false
		node.SubdivideWithPool(pool)

		idxOld := node.getOctantIndex(existingBody.Position)
		node.Children[idxOld].insertWithPool(existingBody, depth+1, pool)

		idxNew := node.getOctantIndex(body.Position)
		node.Children[idxNew].insertWithPool(body, depth+1, pool)

		newMass := node.TotalMass + body.Mass
		if newMass > 0 {
			invM := 1.0 / newMass
			node.CenterOfMass.X = float32((float64(node.CenterOfMass.X)*node.TotalMass + float64(body.Position.X)*body.Mass) * invM)
			node.CenterOfMass.Y = float32((float64(node.CenterOfMass.Y)*node.TotalMass + float64(body.Position.Y)*body.Mass) * invM)
			node.CenterOfMass.Z = float32((float64(node.CenterOfMass.Z)*node.TotalMass + float64(body.Position.Z)*body.Mass) * invM)
		}
		node.TotalMass = newMass
		node.Count++
		return
	}

	// 3. Internal node: insert into child octant and update CoM
	idx := node.getOctantIndex(body.Position)
	node.Children[idx].insertWithPool(body, depth+1, pool)

	newMass := node.TotalMass + body.Mass
	if newMass > 0 {
		invM := 1.0 / newMass
		node.CenterOfMass.X = float32((float64(node.CenterOfMass.X)*node.TotalMass + float64(body.Position.X)*body.Mass) * invM)
		node.CenterOfMass.Y = float32((float64(node.CenterOfMass.Y)*node.TotalMass + float64(body.Position.Y)*body.Mass) * invM)
		node.CenterOfMass.Z = float32((float64(node.CenterOfMass.Z)*node.TotalMass + float64(body.Position.Z)*body.Mass) * invM)
	}
	node.TotalMass = newMass
	node.Count++
}

// ComputeAcceleration computes the gravitational acceleration exerted by this node tree on target body
func (root *OctreeNode) ComputeAcceleration(target *Body, g float64, softening float64, theta float32) rl.Vector3 {
	if root == nil || root.Count == 0 || root.TotalMass <= 0 {
		return rl.NewVector3(0, 0, 0)
	}

	eps2 := float32(softening * softening)
	thetaSq := theta * theta
	gFloat := float32(g)

	var acc rl.Vector3
	var stack [96]*OctreeNode
	stackTop := 0
	stack[0] = root

	for stackTop >= 0 {
		node := stack[stackTop]
		stackTop--

		dx := node.CenterOfMass.X - target.Position.X
		dy := node.CenterOfMass.Y - target.Position.Y
		dz := node.CenterOfMass.Z - target.Position.Z

		distSq := dx*dx + dy*dy + dz*dz + eps2
		if distSq < 0.00001 {
			continue
		}

		if node.IsLeaf {
			if node.Body != nil && node.Body.ID == target.ID {
				continue
			}
			dist := float32(math.Sqrt(float64(distSq)))
			invDist3 := 1.0 / (distSq * dist)
			f := gFloat * float32(node.TotalMass) * invDist3
			acc.X += dx * f
			acc.Y += dy * f
			acc.Z += dz * f
			continue
		}

		cellWidth := node.HalfSize * 2.0
		// Barnes-Hut opening angle criterion: (cellWidth / dist) < theta <=> cellWidth^2 < theta^2 * distSq
		if (cellWidth * cellWidth) < (thetaSq * distSq) {
			dist := float32(math.Sqrt(float64(distSq)))
			invDist3 := 1.0 / (distSq * dist)
			f := gFloat * float32(node.TotalMass) * invDist3
			acc.X += dx * f
			acc.Y += dy * f
			acc.Z += dz * f
			continue
		}

		for i := 0; i < 8; i++ {
			child := node.Children[i]
			if child != nil && child.Count > 0 {
				if stackTop < 95 {
					stackTop++
					stack[stackTop] = child
				}
			}
		}
	}

	return acc
}

// MaxDepth calculates the maximum tree depth
func (node *OctreeNode) MaxDepth() int {
	if node == nil || node.IsLeaf {
		return 1
	}
	maxD := 0
	for i := 0; i < 8; i++ {
		if node.Children[i] != nil {
			d := node.Children[i].MaxDepth()
			if d > maxD {
				maxD = d
			}
		}
	}
	return maxD + 1
}

// BuildOctree constructs a Barnes-Hut octree enclosing all active bodies
func BuildOctree(bodies []*Body) *OctreeNode {
	globalNodePool.Reset()
	return BuildOctreeWithPool(bodies, globalNodePool)
}

// BuildOctreeWithPool constructs a Barnes-Hut octree using node pool
func BuildOctreeWithPool(bodies []*Body, pool *NodePool) *OctreeNode {
	n := len(bodies)
	if n == 0 {
		return nil
	}

	// Calculate 3D bounding box
	minX, maxX := bodies[0].Position.X, bodies[0].Position.X
	minY, maxY := bodies[0].Position.Y, bodies[0].Position.Y
	minZ, maxZ := bodies[0].Position.Z, bodies[0].Position.Z

	for _, b := range bodies[1:] {
		if b.Position.X < minX {
			minX = b.Position.X
		}
		if b.Position.X > maxX {
			maxX = b.Position.X
		}
		if b.Position.Y < minY {
			minY = b.Position.Y
		}
		if b.Position.Y > maxY {
			maxY = b.Position.Y
		}
		if b.Position.Z < minZ {
			minZ = b.Position.Z
		}
		if b.Position.Z > maxZ {
			maxZ = b.Position.Z
		}
	}

	centerX := (minX + maxX) * 0.5
	centerY := (minY + maxY) * 0.5
	centerZ := (minZ + maxZ) * 0.5

	spanX := maxX - minX
	spanY := maxY - minY
	spanZ := maxZ - minZ

	maxSpan := spanX
	if spanY > maxSpan {
		maxSpan = spanY
	}
	if spanZ > maxSpan {
		maxSpan = spanZ
	}

	halfSize := (maxSpan * 0.5) + 5.0
	if halfSize < 10.0 {
		halfSize = 10.0
	}

	var root *OctreeNode
	center := rl.NewVector3(centerX, centerY, centerZ)
	if pool != nil {
		root = pool.Alloc(center, halfSize)
	} else {
		root = NewOctreeNode(center, halfSize)
	}

	for _, b := range bodies {
		root.insertWithPool(b, 0, pool)
	}

	return root
}

// CalculateAccelerationsBarnesHut computes accelerations for all bodies using the parallel Barnes-Hut Octree
func CalculateAccelerationsBarnesHut(bodies []*Body, g float64, softening float64, theta float32) []rl.Vector3 {
	n := len(bodies)
	accs := make([]rl.Vector3, n)
	if n == 0 {
		return accs
	}

	globalNodePool.Reset()
	root := BuildOctreeWithPool(bodies, globalNodePool)
	if root == nil {
		return accs
	}

	if theta <= 0.05 {
		theta = 0.7
	}

	numWorkers := runtime.GOMAXPROCS(0)
	if numWorkers < 1 {
		numWorkers = 1
	}

	if n < 120 || numWorkers == 1 {
		for i := 0; i < n; i++ {
			b := bodies[i]
			if b.IsStationary {
				continue
			}
			acc := root.ComputeAcceleration(b, g, softening, theta)
			accs[i] = acc
			bodies[i].NetForce = rl.NewVector3(
				acc.X*float32(b.Mass),
				acc.Y*float32(b.Mass),
				acc.Z*float32(b.Mass),
			)
		}
		return accs
	}

	chunkSize := (n + numWorkers - 1) / numWorkers
	var wg sync.WaitGroup

	for w := 0; w < numWorkers; w++ {
		start := w * chunkSize
		end := start + chunkSize
		if end > n {
			end = n
		}
		if start >= end {
			break
		}

		wg.Add(1)
		go func(s, e int) {
			defer wg.Done()
			for i := s; i < e; i++ {
				b := bodies[i]
				if b.IsStationary {
					continue
				}
				acc := root.ComputeAcceleration(b, g, softening, theta)
				accs[i] = acc
				bodies[i].NetForce = rl.NewVector3(
					acc.X*float32(b.Mass),
					acc.Y*float32(b.Mass),
					acc.Z*float32(b.Mass),
				)
			}
		}(start, end)
	}

	wg.Wait()
	return accs
}
