package util

import (
	"bytes"
	"encoding/json"
	"errors"
	"io"
	"log"
	"os"
	"path/filepath"
	"runtime/debug"
	"strconv"
	"strings"
	"time"
)

// PE Panic on error
func PE(err error) {
	if err != nil {
		panic(err)
	}
}

type Errable func() error

// Warning: later panic in defered call shadows the previous error.
// Becareful when using with defer
func PEF(errable Errable) {
	PE(errable())
}

func Debug(_ interface{}) {
}

// ignore error
func IE(err error) {
	if err != nil {
		log.Println(err)
	}
}

// Panic errors!
func Jsonify(v interface{}) string {
	b, err := json.Marshal(v)
	PE(err)
	return string(b)
}

func LogFilePath() {
	// log.SetFlags(log.LstdFlags | log.Lshortfile)
	log.SetFlags(log.LstdFlags | log.Llongfile)
}

// Catch panic, and print stack info
// set err if nil
func PanicCatcher(err *error) {
	if r := recover(); r != nil {
		log.Printf("Error: %s\nstacktrace from panic: %s\n", r, debug.Stack())
		if e := r.(error); e != nil && *err != e {
			if *err != nil {
				log.Println("Error be shadowed:", e, "For:", *err)
			} else {
				*err = e
			}
		}
	}
}

func DatestrToTime(s string) (t time.Time, err error) {
	err = errors.New("Invalid datestr format")
	segs := strings.Split(s, "/")
	if len(segs) != 3 {
		return
	}
	y, err := strconv.Atoi(segs[0])
	m, err := strconv.Atoi(segs[1])
	d, err := strconv.Atoi(segs[2])
	if err != nil {
		return
	}
	t = time.Date(y, time.Month(m), d, 0, 0, 0, 0, time.UTC)
	return
}

// DateFromStr convert datestr in form "2020/01/02"(UTC) to timestamp
func DatestrToInt(s string) (timestamp int64, err error) {
	t, err := DatestrToTime(s)
	timestamp = t.Unix()
	return
}

func DatestrFromInt(t int) (datestr string) {
	return
}

func CopyFile(dst, src string) (sz int64, err error) {
	in, err := os.Open(src)
	if err != nil {
		return
	}
	defer in.Close()

	out, err := os.Create(dst) // behold
	if err != nil {
		return
	}
	defer out.Close()

	sz, err = io.Copy(out, in)
	if err != nil {
		return
	}
	return
}
func CopyFileN(dst, src string, n int64) (sz int64, err error) {
	in, err := os.Open(src)
	if err != nil {
		return
	}
	defer in.Close()

	out, err := os.Create(dst) // behold
	if err != nil {
		return
	}
	defer out.Close()

	sz, err = io.CopyN(out, in, n)
	if err != nil {
		return
	}
	return
}

type IndexedReadCloser struct {
	R   io.Reader
	C   io.Closer
	Idx int64
	Len int64
}

// Read reads and counts.
func (r *IndexedReadCloser) Read(buf []byte) (n int, err error) {
	n, err = r.R.Read(buf)
	r.Idx += int64(n)
	return
}

func (r *IndexedReadCloser) Close() error {
	return r.C.Close()
}

func (r *IndexedReadCloser) Rest() int64 {
	return r.Len - r.Idx
}

func (r *IndexedReadCloser) EOF() bool {
	return r.Idx >= r.Len
}

func EnsureDir(p string) {
	fi, err := os.Stat(p)
	if err == nil {
		if fi.IsDir() {
			return
		}
		panic(errors.New("path is already a file: " + p))
	}
	if os.IsNotExist(err) {
		os.Mkdir(p, 0775)
	} else {
		panic(err)
	}
}

func ReportErr(err error) {
	if err != nil {
		log.Println(err)
	}
}

// Assert panic on false conf
func Assert(cond bool, why string) {
	if !cond {
		panic(errors.New("Assertion failed: \n" + why))
	}
}

func FatalIfErr(err error) {
	if err != nil {
		log.Fatal(err)
	}
}

func Min(x, y int64) int64 {
	if x < y {
		return x
	}
	return y
}

// call clean up function and return if error, panic on overlapped error
// func CleanUpPanic(errable Errable, err *error) {
// 	e := errable()
// 	if e != nil {
// 		if *err != nil {
// 			panic(e) // panic on multi err
// 		} else {
// 			*err = e
// 		}
// 	}
// }

// call clean up function and return if error, cry and cowered on overlapped error
func CleanUp(errable Errable, err *error) {
	e := errable()
	if e != nil {
		if *err != nil {
			log.Println("Error be shadowed during cleanup:", e, "For:", *err) // panic on multi err
		} else {
			*err = e
		}
	}
}

// CreateFile create a new zero-filled file
func CreateFile(p string, sz int64) (err error) {
	f, err := os.OpenFile(p, os.O_RDWR|os.O_CREATE|os.O_EXCL, 0644)
	if err != nil {
		return
	}
	defer CleanUp(f.Close, &err)

	err = f.Truncate(sz)
	return
}

// OpenFile open file and seek to pos, for write
func OpenFile(p string, pos int64) (f *os.File, err error) {
	f, err = os.OpenFile(p, os.O_RDWR, 0644)
	if err != nil {
		return
	}

	_, err = f.Seek(pos, 0)
	return
}

func ContainsAny(s string, subs []string) bool {
	for _, sub := range subs {
		if strings.Contains(s, sub) {
			return true
		}
	}
	return false
}

func IncludedInAny(sub string, lst []string) bool {
	for _, s := range lst {
		if strings.Contains(s, sub) {
			return true
		}
	}
	return false
}

func IsOneOf(s string, lst []string) bool {
	for _, ss := range lst {
		if s == ss {
			return true
		}
	}
	return false
}

// MakePath under wd
func MakePath(wd, p string) string {
	if filepath.IsAbs(p) {
		return p
	}
	return filepath.Join(wd, p)
}

func IsPathError(err error) bool {
	_, ok := err.(*os.PathError)
	return ok
}

func ReadDir(dirname string) ([]os.FileInfo, error) {
	f, err := os.Open(dirname)
	if err != nil {
		return nil, err
	}
	list, err := f.Readdir(-1)
	f.Close()
	if err != nil {
		return nil, err
	}
	// sort.Slice(list, func(i, j int) bool { return list[i].Name() < list[j].Name() })
	return list, nil
}

func LoopDo(f Errable, intervalMs int, msg string) {
	for {
		if len(msg) > 0 {
			log.Println(msg)
		}
		err := f()
		if err != nil {
			log.Println(err)
			// log.Printf("%s\n", debug.Stack())
		}
		time.Sleep(time.Millisecond * time.Duration(intervalMs))
	}
}

func StrFromSzslice(byteArray []byte) string {
	n := bytes.IndexByte(byteArray, 0)
	if n >= 0 {
		return string(byteArray[:n])
	}
	return string(byteArray[:])
}

func FileExists(p string) bool {
	fi, err := os.Stat(p)
	if err == nil {
		return !fi.IsDir()
	}
	return false
}

func FileSize(p string) (n int64, err error) {
	fi, err := os.Stat(p)
	if err == nil {
		n = fi.Size()
	}
	return
}
