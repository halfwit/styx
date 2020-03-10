package styxfile

import (
	"bytes"
	"io"
	"io/ioutil"
	"os"
	"testing"

	"github.com/go9p/styx/internal/qidpool"
	"github.com/go9p/styx/styxproto"
)

func compare(t *testing.T, file Interface, offset int64, want string) {
	buf := make([]byte, 1000)
	n, err := file.ReadAt(buf, offset)
	if err != nil {
		if n == 0 {
			t.Fatal(err, n)
		}
	}
	got := string(buf[:n])
	if got != want {
		t.Errorf("ReadAt(f, %d) got %q, want %q", offset, got, want)
	} else {
		t.Logf("ReadAt(f, %d) = %q", offset, got)
	}
}

func write(t *testing.T, file Interface, offset int64, data string) {
	_, err := file.WriteAt([]byte(data), offset)
	if err != nil {
		t.Fatal(err)
	}
}

func TestSeeker(t *testing.T) {
	r := bytes.NewReader([]byte("hello, world!"))

	file, err := New(r)
	if err != nil {
		t.Fatal(err)
	}
	defer file.Close()

	compare(t, file, 0, "hello, world!")
	compare(t, file, 1, "ello, world!")
	compare(t, file, 7, "world!")
}

func TestDumb(t *testing.T) {
	var buf bytes.Buffer

	file, err := New(&buf)
	if err != nil {
		t.Fatal(err)
	}
	defer file.Close()

	write(t, file, 0, "h")
	write(t, file, 1, "e")
	write(t, file, 2, "l")
	write(t, file, 3, "l")
	write(t, file, 4, "o")
}

func TestDirectory(t *testing.T) {
	dirname, err := ioutil.TempDir("", "")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(dirname)

	fd, err := os.Open(dirname)
	if err != nil {
		t.Error(err)
		return
	}

	for i := 0; i < 10; i++ {
		f, err := ioutil.TempFile(dirname, "dirtest")
		if err != nil {
			t.Error(err)
		}
		f.Close()
	}
	_, err = ioutil.TempDir(dirname, "dirtest-dir")
	if err != nil {
		t.Error(err)
		return
	}

	dir := NewDir(fd, dirname, qidpool.New())

	// We know that we can read a single Stat by only
	// asking for 1 * MaxStatLen bytes. This is an implementation
	// detail that may not be true in the future.
	buf := make([]byte, styxproto.MaxStatLen)
	var offset int64

	for {
		n, err := dir.ReadAt(buf, offset)
		if err == io.EOF {
			break
		}
		if err != nil {
			t.Error(err)
			break
		}
		offset += int64(n)
		stat := styxproto.Stat(buf)
		t.Logf("%s", stat)
	}
}
