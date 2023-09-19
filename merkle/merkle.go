package merkle

import (
	"bytes"
	"crypto/sha256"
	"errors"
	"log"
)

type Node struct {
	Hash   []byte
	Left   *Node
	Right  *Node
	Parent *Node
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

func (mt *MerkleTree) recalculateTree() error {
	if len(mt.Leaves) == 0 {
		return errors.New("no leaves to build tree")
	}

	nodes := mt.Leaves
	for len(nodes) > 1 {
		var nextLevel []*Node

		for i := 0; i < len(nodes); i += 2 {
			// If we're at the end and there's an odd number of nodes, duplicate the last one.
			if i+1 == len(nodes) {
				nodes = append(nodes, nodes[i])
			}

			// Enforce comparison for ordering
			first, second := nodes[i].Hash, nodes[i+1].Hash
			if bytes.Compare(first, second) > 0 {
				first, second = second, first
			}

			combinedHash := append(first, second...)
			hash := sha256.Sum256(combinedHash)

			newNode := &Node{
				Hash:  hash[:],
				Left:  nodes[i],
				Right: nodes[i+1],
			}
			nodes[i].Parent = newNode
			nodes[i+1].Parent = newNode

			nextLevel = append(nextLevel, newNode)
		}

		nodes = nextLevel
	}

	mt.Root = nodes[0]
	return nil
}

// ComputeRoot returns the Merkle root.
func (mt *MerkleTree) ComputeRoot() (*Node, error) {
	if mt.Root == nil {
		return nil, errors.New("merkle tree has not been calculated yet")
	}
	return mt.Root, nil
}

func (mt *MerkleTree) GenerateProof(leafIndex int) ([][]byte, error) {
	if leafIndex < 0 || leafIndex >= len(mt.Leaves) {
		return nil, errors.New("invalid leaf index")
	}

	var proof [][]byte
	current := mt.Leaves[leafIndex]

	for current.Parent != nil {
		sibling := getSibling(current)
		if sibling != nil {
			log.Printf("sibling: %x", sibling.Hash)
			proof = append(proof, sibling.Hash)
		}
		current = current.Parent
	}

	return proof, nil
}

// getSibling returns the sibling of a given node
func getSibling(node *Node) *Node {
	if node.Parent == nil {
		return nil
	}
	if node.Parent.Left == node {
		return node.Parent.Right
	} else {
		return node.Parent.Left
	}
}

// GetIndexFromContent returns the index of a leaf node given its content.
func (mt *MerkleTree) GetIndexFromContent(content []byte) int {
	hash := sha256.Sum256(content)
	return mt.getIndexFromHash(hash[:])
}

func (mt *MerkleTree) getIndexFromHash(hash []byte) int {
	for i, leaf := range mt.Leaves {
		if bytes.Equal(leaf.Hash, hash) {
			return i
		}
	}
	return -1
}
