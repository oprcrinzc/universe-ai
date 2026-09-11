package main

import "math"

const MaxPoolNodes = 65536

type OctreeNode struct {
	BoundsMin    Vector3
	BoundsMax    Vector3
	Center       Vector3
	HalfWidth    float32
	Mass         float64
	CenterOfMass Vector3
	Body         *Body
	Children     [8]*OctreeNode
	IsLeaf       bool
}

type NodePool struct {
	nodes []OctreeNode
	index int
}

var globalNodePool = &NodePool{
	nodes: make([]OctreeNode, MaxPoolNodes),
	index: 0,
}

func (p *NodePool) Reset() {
	p.index = 0
}

func (p *NodePool) Alloc() *OctreeNode {
	if p.index < len(p.nodes) {
		n := &p.nodes[p.index]
		p.index++
		*n = OctreeNode{IsLeaf: true}
		return n
	}
	return &OctreeNode{IsLeaf: true}
}

type Octree struct {
	Root *OctreeNode
	Pool *NodePool
}

func NewOctree(bodies []*Body) *Octree {
	pool := globalNodePool
	pool.Reset()
	if len(bodies) == 0 {
		return &Octree{Pool: pool}
	}

	minB := bodies[0].Position
	maxB := bodies[0].Position
	for _, b := range bodies[1:] {
		if b.Position.X < minB.X { minB.X = b.Position.X }
		if b.Position.Y < minB.Y { minB.Y = b.Position.Y }
		if b.Position.Z < minB.Z { minB.Z = b.Position.Z }
		if b.Position.X > maxB.X { maxB.X = b.Position.X }
		if b.Position.Y > maxB.Y { maxB.Y = b.Position.Y }
		if b.Position.Z > maxB.Z { maxB.Z = b.Position.Z }
	}

	sizeX := maxB.X - minB.X
	sizeY := maxB.Y - minB.Y
	sizeZ := maxB.Z - minB.Z
	maxDim := sizeX
	if sizeY > maxDim { maxDim = sizeY }
	if sizeZ > maxDim { maxDim = sizeZ }
	if maxDim < 10.0 { maxDim = 10.0 }
	maxDim *= 1.1

	center := Vector3Scale(Vector3Add(minB, maxB), 0.5)
	halfW := maxDim * 0.5
	bMin := Vector3{X: center.X - halfW, Y: center.Y - halfW, Z: center.Z - halfW}
	bMax := Vector3{X: center.X + halfW, Y: center.Y + halfW, Z: center.Z + halfW}

	root := pool.Alloc()
	root.BoundsMin = bMin
	root.BoundsMax = bMax
	root.Center = center
	root.HalfWidth = halfW

	tree := &Octree{Root: root, Pool: pool}
	for _, b := range bodies {
		tree.Insert(root, b, 0)
	}
	return tree
}

func (t *Octree) Insert(node *OctreeNode, b *Body, depth int) {
	if depth > 32 {
		node.Mass += b.Mass
		return
	}

	if node.Mass == 0 && node.Body == nil && node.IsLeaf {
		node.Body = b
		node.Mass = b.Mass
		node.CenterOfMass = b.Position
		return
	}

	if node.IsLeaf && node.Body != nil {
		oldBody := node.Body
		node.Body = nil
		node.IsLeaf = false
		t.subdivide(node)

		t.insertToChild(node, oldBody, depth+1)
		t.insertToChild(node, b, depth+1)

		node.Mass = oldBody.Mass + b.Mass
		totalM := float32(node.Mass)
		node.CenterOfMass = Vector3{
			X: (oldBody.Position.X*float32(oldBody.Mass) + b.Position.X*float32(b.Mass)) / totalM,
			Y: (oldBody.Position.Y*float32(oldBody.Mass) + b.Position.Y*float32(b.Mass)) / totalM,
			Z: (oldBody.Position.Z*float32(oldBody.Mass) + b.Position.Z*float32(b.Mass)) / totalM,
		}
		return
	}

	if !node.IsLeaf {
		t.insertToChild(node, b, depth+1)
		newMass := node.Mass + b.Mass
		ratioOld := float32(node.Mass / newMass)
		ratioNew := float32(b.Mass / newMass)
		node.CenterOfMass = Vector3{
			X: node.CenterOfMass.X*ratioOld + b.Position.X*ratioNew,
			Y: node.CenterOfMass.Y*ratioOld + b.Position.Y*ratioNew,
			Z: node.CenterOfMass.Z*ratioOld + b.Position.Z*ratioNew,
		}
		node.Mass = newMass
	}
}

func (t *Octree) subdivide(node *OctreeNode) {
	qW := node.HalfWidth * 0.5
	for i := 0; i < 8; i++ {
		ox, oy, oz := -qW, -qW, -qW
		if (i & 1) != 0 { ox = qW }
		if (i & 2) != 0 { oy = qW }
		if (i & 4) != 0 { oz = qW }

		childCenter := Vector3{X: node.Center.X + ox, Y: node.Center.Y + oy, Z: node.Center.Z + oz}
		child := t.Pool.Alloc()
		child.Center = childCenter
		child.HalfWidth = qW
		child.BoundsMin = Vector3{X: childCenter.X - qW, Y: childCenter.Y - qW, Z: childCenter.Z - qW}
		child.BoundsMax = Vector3{X: childCenter.X + qW, Y: childCenter.Y + qW, Z: childCenter.Z + qW}
		node.Children[i] = child
	}
}

func (t *Octree) getOctant(node *OctreeNode, pos Vector3) int {
	oct := 0
	if pos.X >= node.Center.X { oct |= 1 }
	if pos.Y >= node.Center.Y { oct |= 2 }
	if pos.Z >= node.Center.Z { oct |= 4 }
	return oct
}

func (t *Octree) insertToChild(node *OctreeNode, b *Body, depth int) {
	oct := t.getOctant(node, b.Position)
	t.Insert(node.Children[oct], b, depth)
}

func (t *Octree) ComputeAcceleration(target *Body, G, softening, theta float64) (Vector3, int64) {
	if t.Root == nil || t.Root.Mass <= 0 {
		return Vector3{}, 0
	}

	softSq := float32(softening * softening)
	thetaSq := float32(theta * theta)
	acc := Vector3{}
	var interactions int64 = 0

	var stack [64]*OctreeNode
	stackPtr := 0
	stack[0] = t.Root
	stackPtr = 1

	for stackPtr > 0 {
		stackPtr--
		curr := stack[stackPtr]
		if curr == nil || curr.Mass <= 0 {
			continue
		}

		if curr.IsLeaf {
			if curr.Body != nil && curr.Body.ID != target.ID {
				interactions++
				dx := curr.Body.Position.X - target.Position.X
				dy := curr.Body.Position.Y - target.Position.Y
				dz := curr.Body.Position.Z - target.Position.Z
				distSq := dx*dx + dy*dy + dz*dz + softSq
				invDist := 1.0 / float32(math.Sqrt(float64(distSq)))
				invDist3 := invDist / distSq
				f := float32(G*curr.Body.Mass) * invDist3
				acc.X += dx * f
				acc.Y += dy * f
				acc.Z += dz * f
			}
			continue
		}

		dx := curr.CenterOfMass.X - target.Position.X
		dy := curr.CenterOfMass.Y - target.Position.Y
		dz := curr.CenterOfMass.Z - target.Position.Z
		distSq := dx*dx + dy*dy + dz*dz + softSq
		cellW := curr.HalfWidth * 2.0

		if (cellW * cellW) < (thetaSq * distSq) {
			interactions++
			invDist := 1.0 / float32(math.Sqrt(float64(distSq)))
			invDist3 := invDist / distSq
			f := float32(G*curr.Mass) * invDist3
			acc.X += dx * f
			acc.Y += dy * f
			acc.Z += dz * f
		} else {
			for i := 0; i < 8; i++ {
				ch := curr.Children[i]
				if ch != nil && ch.Mass > 0 {
					if stackPtr < len(stack) {
						stack[stackPtr] = ch
						stackPtr++
					}
				}
			}
		}
	}
	return acc, interactions
}
