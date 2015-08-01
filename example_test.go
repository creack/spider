package spider_test

import (
	"fmt"
	"log"
	"time"

	"github.com/creack/spider"
)

func Example() {
	var err error
	_ = err
	n := spider.NewNode()
	_ = n.Create("/a/b/c/ff/gg", 42)
	_ = n.Create("/a/b/d", []byte("hello"))
	_ = n.Create("/map", map[string]interface{}{"foo": "bar", "foo3": []int{1, 2}, "foo2": map[string]interface{}{"subfoo": "subbar"}})
	_ = n.Create("/foo/slice", []interface{}{"a", 2, []string{"foo", "bar"}, nil, map[int]time.Time{42: time.Now()}})
	_ = n.Create("/map/foo3", "ok")

	// if err != nil {
	// 	fmt.Println(err)
	// } else {
	// 	n.DumpTree()
	// }

	nn, err := n.Get("/map/foo3")
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println(nn.Path)
	fmt.Println(nn)
}
