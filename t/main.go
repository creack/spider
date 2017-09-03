package main

import (
	"encoding/json"
	"fmt"
	"log"
	"time"

	"github.com/creack/spider"
)

func main() {
	var err error
	_ = err
	n := spider.NewNode()
	if err := n.Create("/a/b/c/ff/gg", 42); err != nil {
		log.Fatal(err)
	}
	if err := n.Create("/a/b/d", "hello"); err != nil {
		log.Fatal(err)
	}
	if err := n.Create("/map", map[string]interface{}{"foo": "bar", "foo3": []int{1, 2}, "foo2": map[string]interface{}{"subfoo": "subbar"}}); err != nil {
		log.Fatal(err)
	}
	if err := n.Create("/foo/slice", []interface{}{"a", 2, []string{"foo", "bar"}, nil, map[int]time.Time{42: time.Now()}}); err != nil {
		log.Fatal(err)
	}
	// fmt.Printf("->%#v\n", n.Children)
	// if err != nil {
	// 	fmt.Println(err)
	// } else {
	// 	n.DumpTree()
	// }

	nn, err := n.Get("/map/foo")
	if err != nil {
		log.Fatal(err)
	}
	_ = nn
	buf, err := json.MarshalIndent(nn, "", "  ")
	if err != nil {
		log.Fatal("->", err)
	}
	_ = buf
	fmt.Printf("%s\n", nn)
}
