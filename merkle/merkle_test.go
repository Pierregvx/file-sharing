package merkle

import (
	"bytes"
	"crypto/sha256"
	"testing"
)

func TestAddLeaves(t *testing.T) {
	mt := NewMerkleTree()
	leaves := [][]byte{{'a'}, {'b'}, {'c'}, {'d'}}

	mt.AddLeaves(leaves)

	if len(mt.Leaves) != 4 {
		t.Errorf("Expected 4 leaves, but got %d leaves", len(mt.Leaves))
	}
}

func TestAddFile(t *testing.T) {
	mt := NewMerkleTree()
	content := []byte{'e'}
	err := mt.AddFile(content)

	if err != nil {
		t.Errorf("Failed to add file: %v", err)
	}
	hash := sha256.Sum256(content)
	if len(mt.Leaves) != 1 || !bytes.Equal(mt.Leaves[0].Hash,hash[:] ) {
		t.Errorf("Leaf was not added correctly")
	}
}

func TestRecalculateTree(t *testing.T) {
	mt := NewMerkleTree()
	leaves := [][]byte{{'a'}, {'b'}, {'c'}, {'d'}}
	mt.AddLeaves(leaves)

	err := mt.recalculateTree()
	if err != nil {
		t.Errorf("Failed to recalculate tree: %v", err)
	}

	if mt.Root == nil {
		t.Error("Root should not be nil after recalculation")
	}
}

func TestComputeRoot(t *testing.T) {
	mt := NewMerkleTree()
	leaves := [][]byte{{'a'}, {'b'}, {'c'}, {'d'}}
	mt.AddLeaves(leaves)

	_, err := mt.ComputeRoot()
	if err != nil {
		t.Errorf("Failed to compute root: %v", err)
	}
}

func TestGenerateProof(t *testing.T) {
	mt := NewMerkleTree()
	leaves := [][]byte{{'a'}, {'b'}, {'c'}, {'d'}}
	mt.AddLeaves(leaves)

	proof, err := mt.GenerateProof(1) // For leaf 'b'
	if err != nil {
		t.Fatalf("Failed to generate proof: %v", err)
	}

	if len(proof) == 0 {
		t.Error("Proof should not be empty")
	}
}