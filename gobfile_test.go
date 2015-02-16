package gobfile

import (
	"fmt"
	"io/ioutil"
	"math/rand"
	"os"
	"path/filepath"
	"strconv"
	"strings"
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
	file, err := New(&obj, path, NewPortLocker(port))
	if err != nil {
		t.Fatalf("new %v", err)
	}
	err = file.Save()
	if err != nil {
		t.Fatalf("save %v", err)
	}
	file.Close()

	var obj2 Object
	file, err = New(&obj2, path, NewPortLocker(port))
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

func TestInvalidObject(t *testing.T) {
	_, err := New(42, "foo", NewPortLocker(0))
	if err == nil || err.Error() != "object must be a pointer" {
		t.Fatal("should fail")
	}
}

func TestLockFail(t *testing.T) {
	port := rand.Intn(20000) + 30000
	path := filepath.Join(os.TempDir(), "gobfile-test-"+strconv.FormatInt(rand.Int63(), 10))
	obj := struct{}{}
	_, err := New(&obj, path, NewPortLocker(port))
	if err != nil {
		t.Fatal(err)
	}

	_, err = New(&obj, path, NewPortLocker(port))
	if err == nil || err.Error() != "lock fail" {
		t.Fatal("should fail")
	}
}

func TestCorruptedFile(t *testing.T) {
	obj := map[int]string{
		1: "foo",
		2: "bar",
		3: "baz",
	}
	port := rand.Intn(20000) + 30000
	path := filepath.Join(os.TempDir(), "gobfile-test-"+strconv.FormatInt(rand.Int63(), 10))
	file, err := New(&obj, path, NewPortLocker(port))
	if err != nil {
		t.Fatal(err)
	}
	file.Save()
	file.Close()

	content, err := ioutil.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	content = content[:len(content)/2]
	err = ioutil.WriteFile(path, content, 0644)
	if err != nil {
		t.Fatal(err)
	}
	file, err = New(&obj, path, NewPortLocker(port))
	if err == nil || err.Error() != "gob file decode error: unexpected EOF" {
		t.Fatal("should fail")
	}
}

func TestSaveFail(t *testing.T) {
	dir := filepath.Join(os.TempDir(), fmt.Sprintf("gobfile-testdir-%d", rand.Int63()))
	os.Mkdir(dir, 0755)
	path := filepath.Join(dir, "foo")
	port := rand.Intn(20000) + 30000
	obj := struct{}{}
	file, err := New(&obj, path, NewPortLocker(port))
	if err != nil {
		t.Fatal(err)
	}
	os.Chmod(dir, 0000)
	err = file.Save()
	if err == nil || !strings.HasPrefix(err.Error(), "open temp file error:") {
		t.Fatal("should fail")
	}
}

func TestEncodeError(t *testing.T) {
	port := rand.Intn(20000) + 30000
	path := filepath.Join(os.TempDir(), "gobfile-test-"+strconv.FormatInt(rand.Int63(), 10))
	obj := func() {}
	file, err := New(&obj, path, NewPortLocker(port))
	if err != nil {
		t.Fatal(err)
	}
	err = file.Save()
	if err == nil || err.Error() != "gob NewTypeObject can't handle type: func()" {
		t.Fatal("should fail")
	}
}

func TestSaveFail2(t *testing.T) {
	port := rand.Intn(20000) + 30000
	path := filepath.Join(os.TempDir(), "gobfile-test-"+strconv.FormatInt(rand.Int63(), 10))
	obj := struct{}{}
	file, err := New(&obj, path, NewPortLocker(port))
	if err != nil {
		t.Fatal(err)
	}
	os.Remove(path)
	os.Mkdir(path, 0000)
	err = file.Save()
	if err == nil || !strings.HasPrefix(err.Error(), "temp file rename error:") {
		t.Fatal("should fail")
	}
}
