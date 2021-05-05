// remove go1.16 dependency, build go1.16

package gnet

import (
	"testing"
	"io"
	"os"
	"fmt"
	// "io/fs"
)

func TestFSGet(t *testing.T) {
	fp := Get("http://httpbin.org/get")
	fs_output(fp)
	fmt.Printf("\n---- done to TestFSGet() ---\n\n")
}

func TestFSJson(t *testing.T) {
	fp := Post("http://httpbin.org/post", JSONCall(), Params(map[string]interface{}{"a": "b", "c": 1})) 
	fs_output(fp)
	fmt.Printf("\n---- done to TestFSJson() ---\n\n")
}

func fs_output(fp *File) {
	defer fp.Close()
	io.Copy(os.Stdout, fp)
}

func TestFSParseJSON(t *testing.T) {
	var res map[string]interface{}
	status, err := FsCallAndParseJSON("http://httpbin.org/post", "POST", &res, Params(map[string]interface{}{"a": "b", "c": 1}), JSONCall(), BodyLogger(os.Stderr))
	if err != nil {
		fmt.Printf("failed to call FsCallAndParseJSON: %v\n", err)
		return
	}
	fmt.Printf("status: %d, res: %#v\n", status, res)
	fmt.Printf("\n---- done to TestFSParseJSON() ---\n\n")
}

