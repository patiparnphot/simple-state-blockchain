package merkletrie

import (
	"bytes"
	"crypto/sha256"
	"fmt"
	"runtime"
)

type MerkleTrie struct {
	Root       *Node
	MerkleHash []byte
	Leafs      []*Node
}

type Node struct {
	Trie    *MerkleTrie
	Parent  *Node
	Left    *Node
	Right   *Node
	IsLeaf  bool
	IsDup   bool
	Hash    []byte
	Account *Account
}

func NewTrie(accountList []*Account) *MerkleTrie {
	trie := &MerkleTrie{}

	root, leafs := BuildWithAccountList(accountList, trie)

	trie.Root = root
	trie.Leafs = leafs
	trie.MerkleHash = root.Hash

	return trie
}

func (trie *MerkleTrie) ReconstructTrie(accountList []*Account) {
	root, leafs := BuildWithAccountList(accountList, trie)

	trie.Root = root
	trie.Leafs = leafs
	trie.MerkleHash = root.Hash
}

func BuildWithAccountList(accountList []*Account, trie *MerkleTrie) (*Node, []*Node) {
	if len(accountList) == 0 {
		fmt.Println("Cannot build a trie without any account, Create One!!!")
		runtime.Goexit()
	}

	var leafs []*Node

	for _, account := range accountList {
		hash := account.CalHash()

		leafs = append(leafs, &Node{
			Trie:    trie,
			IsLeaf:  true,
			IsDup:   false,
			Account: account,
			Hash:    hash,
		})
	}

	if len(leafs)%2 == 1 {
		duplicate := &Node{
			Trie:    trie,
			IsLeaf:  true,
			IsDup:   true,
			Account: leafs[len(leafs)-1].Account,
			Hash:    leafs[len(leafs)-1].Hash,
		}

		leafs = append(leafs, duplicate)
	}

	root := BuildIntermediate(leafs, trie)

	return root, leafs
}

func BuildIntermediate(leafNodes []*Node, trie *MerkleTrie) *Node {
	var parentNodes []*Node

	for i := 0; i < len(leafNodes); i += 2 {
		var leftIdx, rightIdx int = i, i + 1

		if rightIdx == len(leafNodes) {
			rightIdx = i
		}

		b4hash := bytes.Join(
			[][]byte{
				leafNodes[leftIdx].Hash,
				leafNodes[rightIdx].Hash,
			},
			[]byte{},
		)
		hash := sha256.Sum256(b4hash)

		parentNode := &Node{
			Trie:   trie,
			IsLeaf: false,
			IsDup:  false,
			Left:   leafNodes[leftIdx],
			Right:  leafNodes[rightIdx],
			Hash:   hash[:],
		}

		parentNodes = append(parentNodes, parentNode)

		leafNodes[leftIdx].Parent = parentNode
		leafNodes[rightIdx].Parent = parentNode

		if len(leafNodes) == 2 {
			return parentNode
		}
	}

	return BuildIntermediate(parentNodes, trie)
}

func (node *Node) CalNodeHash() []byte {
	if node.IsLeaf {
		return node.Hash
	}

	b4hash := bytes.Join(
		[][]byte{
			node.Left.Hash,
			node.Right.Hash,
		},
		[]byte{},
	)

	hash := sha256.Sum256(b4hash)

	return hash[:]
}

func (trie *MerkleTrie) VertifyAccount(account Account) bool {
	for _, leaf := range trie.Leafs {
		isEqual := leaf.Account.Equal(account)

		if isEqual {
			currentParent := leaf.Parent

			for currentParent != nil {
				leftHash := currentParent.Left.CalNodeHash()
				rightHash := currentParent.Right.CalNodeHash()

				b4hash := bytes.Join(
					[][]byte{
						leftHash,
						rightHash,
					},
					[]byte{},
				)
				hash := sha256.Sum256(b4hash)

				if !bytes.Equal(hash[:], currentParent.Hash) {
					return false
				}

				currentParent = currentParent.Parent
			}

			return true
		}
	}
	return false
}

func (trie *MerkleTrie) ListAccount() []*Account {
	var accountList []*Account

	for _, leaf := range trie.Leafs {
		if !leaf.IsDup {
			accountList = append(accountList, leaf.Account)
		}
	}

	return accountList
}
