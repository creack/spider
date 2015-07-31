package spider

import (
	"fmt"
	gopath "path"
	"reflect"
	"strconv"
	"strings"
)

// LocalProvider provides an in-memory configuration tree.
type LocalProvider Node

// NewLocalProvider creates a root node and returns it.
func NewLocalProvider() *LocalProvider {
	return &LocalProvider{
		Path:     "/",
		Children: []*Node{},
	}
}

// Get fetches a node from the provider.
func (p *LocalProvider) Get(path string) *Node {
	return p.get(path, (*Node)(p), 1)
}

func (p *LocalProvider) get(path string, n *Node, depth int) *Node {
	parts := strings.SplitN(path, "/", depth)
	if len(parts) < 2 {
		return n
	}
	return p.get(parts[1], n, depth+1)
}

// createEmpty constructs the tree for the given path.
func createEmpty(path string, node *Node, depth int) (*Node, error) {
	// If the current node has the expected path,
	// everything is already created, do nothing.
	if "/"+path == node.Path {
		return node, nil
	}

	// Lookup the direct child path
	parts := strings.SplitN(path, "/", depth+1)
	if len(parts) == 0 {
		return nil, fmt.Errorf("invalid path")
	}
	childPath := "/" + strings.Join(parts[:depth], "/")

	// Check if the child exists
	for _, child := range node.Children {
		// If it does, go to next level.
		if childPath == child.Path {
			return createEmpty(path, child, depth+1)
		}
	}
	if node.Data != nil {
		return nil, fmt.Errorf("Node is a leaf, can't create children")
	}
	// Create the new child
	newChild := &Node{
		Path:     childPath,
		Children: []*Node{},
		Parent:   node,
	}
	node.Children = append(node.Children, newChild)

	// Go to next level.
	return createEmpty(path, newChild, depth+1)
}

// Create adds or set the given data to the path.
// This will create all sub-nodes as necessary.
// Returns the newly created node.
func (p *LocalProvider) Create(path string, data Data) error {
	path = TrimSlash(path)
	newNode, err := createEmpty(path, (*Node)(p), 1)
	if err != nil {
		return err
	}
	switch val := reflect.ValueOf(data); val.Kind() {
	case reflect.Chan:
		return fmt.Errorf("unsuppoted typed: channel")
	case reflect.Map:
		for _, key := range val.MapKeys() {
			p.Create(gopath.Join(path, key.String()), val.MapIndex(key).Interface())
		}
	case reflect.Slice, reflect.Array:
		for i := 0; i < val.Len(); i++ {
			p.Create(gopath.Join(path, strconv.FormatInt(int64(i), 10)), val.Index(i).Interface())
		}
	default:
		if len(newNode.Children) != 0 {
			return fmt.Errorf("node has children, can't set data")
		}
		newNode.Data = data
	}
	return nil
}

func dumpTree(node *Node, depth int) {
	spaces := strings.Repeat("  ", depth-1)
	fmt.Printf("%s%q", spaces, "/"+strings.Split(node.Path, "/")[depth-1])
	if node.Data != nil {
		fmt.Printf(": %v\n", node.Data)
		return
	}
	fmt.Printf("\n")
	for _, child := range node.Children {
		dumpTree(child, depth+1)
	}
}

// DumpTree dumps the content of the current tree.
func (p *LocalProvider) DumpTree() {
	dumpTree((*Node)(p), 1)
}
