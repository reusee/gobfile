package gobfile

import (
	"math/rand"
	"os"
	"path/filepath"
	"strconv"
	"testing"
)

func TestGobFile(t *testing.T) {
	type Object struct {
		Str   string
		Int   int64
		Slice []int
	}
	obj := Object{
		Str:   "foobar",
		Int:   42,
		Slice: []int{5, 3, 2, 1, 4},
	}
	path := filepath.Join(os.TempDir(), "gobfile-test-"+strconv.FormatInt(rand.Int63(), 10))
	port := rand.Intn(20000) + 30000
	file, err := New(&obj, path, port)
	if err != nil {
		t.Fatalf("new %v", err)
	}
	err = file.Save()
	if err != nil {
		t.Fatalf("save %v", err)
	}
	file.Close()

	var obj2 Object
	file, err = New(&obj2, path, port)
	if err != nil {
		t.Fatalf("new %v", err)
	}
	defer file.Close()
	if obj2.Str != obj.Str {
		t.Fatalf("str not match")
	}
	if obj2.Int != obj.Int {
		t.Fatalf("int not match")
	}
	if len(obj2.Slice) != len(obj.Slice) {
		t.Fatalf("slice not match")
	}
	for i, n := range obj2.Slice {
		if n != obj.Slice[i] {
			t.Fatalf("slice not match")
		}
	}
}
