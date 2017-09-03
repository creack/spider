package spider

// DataProvider allows alternate implementations of a backing service that provides configuration data. Ex. Zookeeper, Consul, etcd.
type DataProvider interface {
	Get(key string) (*Node, error)                        // Get an existing key's currently set value.
	Set(key string, value []byte) error                   // Change an existing key's value with an optional version. Set -1 if no versioning is required.
	Create(key string, data []byte) error                 // Create a key with data at that node.
	Delete(key string) error                              // Delete a key from the configuration.
	DeleteR(key string) error                             // Delete a key and it's configuration path recursively.
	Watch(key string) (<-chan interface{}, error)         // Watch returns the broadcast channel used for watching changes on a given configuration node.
	WatchChildren(key string) (<-chan interface{}, error) // WatchChildren monitor for changes to the children of a given node.
	List(key string) ([]string, error)                    // List the child nodes stored under this key.
	Exists(key string) (bool, error)                      // Exists checks if a node has data or if the node exists.
}

// Data represent the data for a given node.
type Data interface{}

// NewConfigTree initialize the configuration tree and return the root node.
func NewConfigTree() *Node {
	return &Node{
		Name:     "/",
		Children: []*Node{},
	}
}

// // Create adds or set the given data to the path.
// // This will create all sub-nodes as necessary.
// // `path` is relative to the current node.
// // Returns the newly created node.
// func (n *Node) Create(path string, data Data) *Node {
// 	newNode := &Node{
// 		Path:   gopath.Join(n.Path, path),
// 		Parent: n,
// 	}
// 	switch data.(type) {
// 	case int, int16, int32, int64, uint16, uint32, uint64,
// 		float32, float64, []byte, string:
// 		newNode.Data = data
// 	case []string, [][]byte, []interface{}, []Data,
// 		map[string]string, map[string]interface{}, map[string]Data, map[string][]byte,
// 		map[string]int, map[string]int16, map[string]int32, map[string]int64,
// 		map[string]uint16, map[string]uint32, map[string]uint64:
// 	}
// 	return newNode
// }
