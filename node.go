package spider

import (
	"encoding/json"
	"fmt"
	gopath "path"
	"reflect"
	"strconv"
	"strings"
)

// Node represent the configuration tree.
type Node struct {
	Name     string  `json:"path"`               // Path within the tree
	Data     Data    `json:"data,omitempty"`     // Data of the current node. If not empty, there is no children.
	Children []*Node `json:"children,omitempty"` // Children of the current node. If not empty, there is no Data.
	Parent   *Node   `json:"-"`                  // Parent of the current node. If empty, the node is the root.

	isArray bool // flag to know if we are in a map int or an array
}

// NewNode creates a root node and returns it.
func NewNode() *Node {
	return &Node{
		Name:     "/",
		Children: []*Node{},
	}
}

// Get fetches a node from the provider.
func (n *Node) Get(path string) (*Node, error) {
	if path == "/" && n.Name == "/" {
		return n, nil
	}
	path = TrimSlash(path)
	return n.get(path, 1)
}

func (n *Node) get(path string, depth int) (*Node, error) {
	parts := strings.SplitN(path, "/", depth+1)
	if len(parts) < depth {
		return n, nil
	}
	childPath := "/" + strings.Join(parts[:depth], "/")

	// Check if the child exists
	for _, child := range n.Children {
		// If it does, go to next level.
		if childPath == child.Name {
			return child.get(path, depth+1)
		}
	}
	return nil, fmt.Errorf("node not found")
}

// MarshalJSON implements json.Marshaler interface.
func (n *Node) MarshalJSON() ([]byte, error) {
	if len(n.Children) == 0 {
		return json.Marshal(n.Data)
	}
	var ret string
	if n.isArray {
		ret = "["
	} else {
		ret = "{"
	}
	for _, child := range n.Children {
		buf, err := child.MarshalJSON()
		if err != nil {
			return nil, err
		}
		if n.isArray {
			ret += string(buf) + ","
		} else {
			ret += `"` + gopath.Base(child.Name) + `":` + string(buf) + ","
		}
	}
	if n.isArray {
		ret = strings.TrimSuffix(ret, ",") + "]"
	} else {
		ret = strings.TrimSuffix(ret, ",") + "}"
	}
	return []byte(ret), nil
}

// String returns a json formatted representation of the tree.
func (n *Node) String() string {
	buf, _ := json.Marshal(n)
	return string(buf)
}

// createEmpty constructs the tree for the given path.
func (n *Node) createEmpty2(path string, depth int) (*Node, error) {
	// If the current node has the expected path,
	// everything is already created, do nothing.
	if "/"+path == n.Name {
		return n, nil
	}

	// Lookup the direct child path
	parts := strings.SplitN(path, "/", depth+1)
	if len(parts) < depth {
		return nil, fmt.Errorf("invalid path")
	}
	childPath := "/" + strings.Join(parts[:depth], "/")

	// Check if the child exists
	for _, child := range n.Children {
		// If it does, go to next level.
		if childPath == child.Name {
			return child.createEmpty2(path, depth+1)
		}
	}
	if n.Data != nil {
		return nil, fmt.Errorf("Node is a leaf, can't create children")
	}
	// Create the new child
	newChild := &Node{
		Name:     childPath,
		Children: []*Node{},
		Parent:   n,
	}
	n.Children = append(n.Children, newChild)

	// Go to next level.
	return newChild.createEmpty2(path, depth+1)
}

// func (n *Node) Create(path string, data Data) error {
// 	path = TrimSlash(path)
// 	newNode, err := n.createEmpty(path, 1)
// 	if err != nil {
// 		return err
// 	}
// 	return n.create(path, data)
// }

// Create2 adds or set the given data to the path.
// This will create all sub-nodes as necessary.
// Returns the newly created node.
func (n *Node) Create2(path string, data Data) error {
	path = TrimSlash(path)
	newNode, err := n.createEmpty2(path, 1)
	if err != nil {
		return err
	}
	// TODO: refactor this with a type switch
	switch val := reflect.ValueOf(data); val.Kind() {
	case reflect.Chan:
		return fmt.Errorf("unsuppoted typed: channel")
	case reflect.Map:
		for _, key := range val.MapKeys() {
			if err := n.Create2(gopath.Join(path, fmt.Sprintf("%v", key.Interface())), val.MapIndex(key).Interface()); err != nil {
				return err
			}
		}
		return nil
	case reflect.Slice, reflect.Array:
		// don't consider []byte as an array.
		if _, ok := data.([]byte); ok {
			break
		}
		newNode.isArray = true
		for i := 0; i < val.Len(); i++ {
			if err := n.Create2(gopath.Join(path, strconv.FormatInt(int64(i), 10)), val.Index(i).Interface()); err != nil {
				return err
			}
		}
		return nil
	}
	if len(newNode.Children) != 0 {
		return fmt.Errorf("node has children, can't set data")
	}
	newNode.Data = data
	return nil
}

// Create .
func (n *Node) Create(path string, data Data) error {
	parts := strings.Split(TrimSlash(path), "/")
	_ = parts
	return nil
}
