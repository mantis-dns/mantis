package gravity

import (
	"strings"
	"sync"
	"sync/atomic"

	iradix "github.com/hashicorp/go-immutable-radix/v2"
)

// Engine manages the Gravity blocklist with lock-free concurrent reads.
type Engine struct {
	blockTree atomic.Pointer[iradix.Tree[struct{}]]
	allowTree atomic.Pointer[iradix.Tree[struct{}]]
	mu        sync.Mutex // protects rebuild
}

// NewEngine creates an empty Gravity engine.
func NewEngine() *Engine {
	e := &Engine{}
	empty := iradix.New[struct{}]()
	e.blockTree.Store(empty)
	e.allowTree.Store(empty)
	return e
}

// IsBlocked checks if a domain is on the blocklist and not on the allowlist.
func (e *Engine) IsBlocked(domain string) bool {
	reversed := ReverseDomain(domain)

	// Allowlist checked first.
	if e.matchTree(e.allowTree.Load(), reversed) {
		return false
	}

	return e.matchTree(e.blockTree.Load(), reversed)
}

// matchTree checks if the reversed domain or any parent prefix exists in the tree.
func (e *Engine) matchTree(tree *iradix.Tree[struct{}], reversed string) bool {
	// Exact match.
	if _, ok := tree.Get([]byte(reversed)); ok {
		return true
	}

	// Wildcard/prefix match: check if any prefix of the reversed domain exists.
	// "com.example.ads.sub." should match "com.example.ads." (wildcard *.ads.example.com)
	parts := strings.Split(strings.TrimSuffix(reversed, "."), ".")
	for i := len(parts) - 1; i >= 1; i-- {
		prefix := strings.Join(parts[:i], ".") + "."
		if _, ok := tree.Get([]byte(prefix)); ok {
			return true
		}
	}

	return false
}

// BlockedCount returns the number of blocked domains.
func (e *Engine) BlockedCount() int {
	return e.blockTree.Load().Len()
}

// RebuildFromDomains atomically replaces the block tree with a new set of domains.
func (e *Engine) RebuildFromDomains(domains []string) {
	e.mu.Lock()
	defer e.mu.Unlock()

	txn := iradix.New[struct{}]().Txn()
	for _, d := range domains {
		key := ReverseDomain(d)
		txn.Insert([]byte(key), struct{}{})
	}
	e.blockTree.Store(txn.Commit())
}

// SetAllowRules atomically replaces the allow tree.
func (e *Engine) SetAllowRules(domains []string) {
	e.mu.Lock()
	defer e.mu.Unlock()

	txn := iradix.New[struct{}]().Txn()
	for _, d := range domains {
		key := ReverseDomain(d)
		txn.Insert([]byte(key), struct{}{})
	}
	e.allowTree.Store(txn.Commit())
}

// AddBlockRule adds a single domain to the block tree.
func (e *Engine) AddBlockRule(domain string) {
	e.mu.Lock()
	defer e.mu.Unlock()

	key := ReverseDomain(domain)
	tree := e.blockTree.Load()
	newTree, _, _ := tree.Insert([]byte(key), struct{}{})
	e.blockTree.Store(newTree)
}

// RemoveBlockRule removes a single domain from the block tree.
func (e *Engine) RemoveBlockRule(domain string) {
	e.mu.Lock()
	defer e.mu.Unlock()

	key := ReverseDomain(domain)
	tree := e.blockTree.Load()
	newTree, _, _ := tree.Delete([]byte(key))
	e.blockTree.Store(newTree)
}

// AddAllowRule adds a single domain to the allow tree.
func (e *Engine) AddAllowRule(domain string) {
	e.mu.Lock()
	defer e.mu.Unlock()

	key := ReverseDomain(domain)
	tree := e.allowTree.Load()
	newTree, _, _ := tree.Insert([]byte(key), struct{}{})
	e.allowTree.Store(newTree)
}

// RemoveAllowRule removes a single domain from the allow tree.
func (e *Engine) RemoveAllowRule(domain string) {
	e.mu.Lock()
	defer e.mu.Unlock()

	key := ReverseDomain(domain)
	tree := e.allowTree.Load()
	newTree, _, _ := tree.Delete([]byte(key))
	e.allowTree.Store(newTree)
}
