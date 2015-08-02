package spider

import (
	"log"
	"os"
	"time"

	"github.com/samuel/go-zookeeper/zk"

	"encoding/json"
	"fmt"
	gopath "path"
	"reflect"
	"strconv"
	"strings"
)

var conn *zk.Conn

func init() {
	c, _, err := zk.Connect([]string{os.Getenv("DOCKER_IP")}, 30*time.Second)
	if err != nil {
		log.Fatal(err)
	}
	conn = c
}

type ZK struct {
	Path     string `json:"path"`               // Path within the tree
	Data     Data   `json:"data,omitempty"`     // Data of the current node. If not empty, there is no children.
	Children []*ZK  `json:"children,omitempty"` // Children of the current node. If not empty, there is no Data.
	Parent   *ZK    `json:"-"`                  // Parent of the current node. If empty, the node is the root.

	isArray bool // flag to know if we are in a map int or an array
}

// NewZK creates a root node and returns it.
func NewZK() *ZK {
	return &ZK{
		Path:     "/",
		Children: []*ZK{},
	}
}

// Get fetches a node from the provider.
func (n *ZK) Get(path string) (*ZK, error) {
	if path == "/" && n.Path == "/" {
		val, _, err := conn.Get("/")
		if err != nil {
			return nil, err
		}
		n.Data = val
		return n, nil
	}
	path = TrimSlash(path)
	return n.get(path, 1)
}

func (n *ZK) get(path string, depth int) (*ZK, error) {
	parts := strings.SplitN(path, "/", depth+1)
	childPath := "/" + strings.Join(parts[:depth], "/")
	if len(parts) < depth {
		childPath = TrimByteSuffix(childPath, '/')
		val, _, err := conn.Get(childPath)
		if err != nil {
			return nil, fmt.Errorf("[zk.Get] %s (%s)", childPath, err)
		}
		if err := json.Unmarshal(val, &n.Data); err != nil {
			return nil, err
		}
		return n, nil
	}

	// Check if the child exists
	for _, child := range n.Children {
		// If it does, go to next level.
		if childPath == child.Path {
			return child.get(path, depth+1)
		}
	}
	return nil, fmt.Errorf("node not found")
}

// MarshalJSON implements json.Marshaler interface.
func (n *ZK) MarshalJSON() ([]byte, error) {
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
			ret += `"` + gopath.Base(child.Path) + `":` + string(buf) + ","
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
func (n *ZK) String() string {
	buf, _ := json.Marshal(n)
	return string(buf)
}

// createEmpty constructs the tree for the given path.
func (n *ZK) createEmpty(path string, depth int) (*ZK, error) {
	// If the current node has the expected path,
	// everything is already created, do nothing.
	if "/"+path == n.Path {
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
		if childPath == child.Path {
			return child.createEmpty(path, depth+1)
		}
	}
	if n.Data != nil {
		return nil, fmt.Errorf("ZK is a leaf, can't create children")
	}

	// TODO: fix this race condition.
	exist, _, err := conn.Exists(childPath)
	if err != nil {
		return nil, fmt.Errorf("[zk.Exists] %s", err)
	}
	if !exist {
		if _, err := conn.Create(childPath, nil, 0, zk.WorldACL(zk.PermAll)); err != nil {
			return nil, fmt.Errorf("[zk.Create] %s", err)
		}
	}
	// Create the new child
	newChild := &ZK{
		Path:     childPath,
		Children: []*ZK{},
		Parent:   n,
	}
	n.Children = append(n.Children, newChild)

	// Go to next level.
	return newChild.createEmpty(path, depth+1)
}

// func (n *ZK) Create(path string, data Data) error {
// 	path = TrimSlash(path)
// 	newZK, err := n.createEmpty(path, 1)
// 	if err != nil {
// 		return err
// 	}
// 	return n.create(path, data)
// }

// Create adds or set the given data to the path.
// This will create all sub-nodes as necessary.
// Returns the newly created node.
func (n *ZK) Create(path string, data Data) error {
	path = TrimSlash(path)
	newZK, err := n.createEmpty(path, 1)
	if err != nil {
		return err
	}
	// TODO: refactor this with a type switch
	switch val := reflect.ValueOf(data); val.Kind() {
	case reflect.Chan:
		return fmt.Errorf("unsuppoted typed: channel")
	case reflect.Map:
		for _, key := range val.MapKeys() {
			if err := n.Create(gopath.Join(path, fmt.Sprintf("%v", key.Interface())), val.MapIndex(key).Interface()); err != nil {
				return err
			}
		}
		return nil
	case reflect.Slice, reflect.Array:
		// don't consider []byte as an array.
		if _, ok := data.([]byte); ok {
			break
		}
		newZK.isArray = true
		for i := 0; i < val.Len(); i++ {
			if err := n.Create(gopath.Join(path, strconv.FormatInt(int64(i), 10)), val.Index(i).Interface()); err != nil {
				return err
			}
		}
		return nil
	}
	if len(newZK.Children) != 0 {
		return fmt.Errorf("node has children, can't set data")
	}
	buf, err := json.Marshal(data)
	if err != nil {
		return err
	}
	if _, err := conn.Set(newZK.Path, buf, -1); err != nil {
		return fmt.Errorf("[zk.Set] %s", err)
	}
	newZK.Data = data
	return nil
}
