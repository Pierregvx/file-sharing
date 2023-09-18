package merkle

import (
	"bytes"
	"crypto/sha256"
	"errors"
)

type Node struct {
	Hash  []byte
	Left  *Node
	Right *Node
}

type MerkleTree struct {
	Root   *Node
	Leaves []*Node
}

// NewMerkleTree creates a new MerkleTree.
func NewMerkleTree() *MerkleTree {
	return &MerkleTree{
		Root:   nil,
		Leaves: []*Node{},
	}
}

// AddFile adds a new file (as a leaf node) and re-calculates the tree.
func (mt *MerkleTree) AddFile(fileContent []byte) {
	hash := sha256.Sum256(fileContent)
	newLeaf := &Node{
		Hash: hash[:],
	}

	mt.Leaves = append(mt.Leaves, newLeaf)
	mt.recalculateTree()
}

// recalculateTree re-calculates the MerkleTree after adding or removing leaves.
func (mt *MerkleTree) recalculateTree() {
	var nodes []Node
	for _, leaf := range mt.Leaves {
		nodes = append(nodes, *leaf)
	}

	for len(nodes) > 1 {
		if len(nodes)%2 != 0 {
			nodes = append(nodes, nodes[len(nodes)-1])
		}

		var level []Node
		for i := 0; i < len(nodes); i += 2 {
			hash := append(nodes[i].Hash, nodes[i+1].Hash...)
			newHash := sha256.Sum256(hash)
			newNode := Node{
				Hash:  newHash[:],
				Left:  &nodes[i],
				Right: &nodes[i+1],
			}
			level = append(level, newNode)
		}

		nodes = level
	}

	if len(nodes) == 1 {
		mt.Root = &nodes[0]
	}
}

// ComputeRoot returns the Merkle root.
func (mt *MerkleTree) ComputeRoot() ([]byte, error) {
	if mt.Root == nil {
		return nil, errors.New("merkle tree has not been calculated yet")
	}
	return mt.Root.Hash, nil
}

// GenerateProof returns the Merkle proof for a given leaf index.
func (mt *MerkleTree) GenerateProof(leafIndex int) ([]byte, error) {
	if leafIndex < 0 || leafIndex >= len(mt.Leaves) {
		return nil, errors.New("invalid leaf index")
	}
	return mt.Leaves[leafIndex].Hash, nil
}

// GetIndexFromContent returns the index of a leaf node given its content.
func  (mt *MerkleTree) GetIndexFromContent(content []byte) int {
	hash := sha256.Sum256(content)
	return mt.getIndexFromHash(hash[:])
}
// 
func (mt *MerkleTree) getIndexFromHash(hash []byte) int {
	for i, leaf := range mt.Leaves {
		if bytes.Equal(leaf.Hash, hash) {
			return i
		}
	}
	return -1
}


