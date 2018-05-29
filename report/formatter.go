package report

import (
	"github.com/spf13/afero"

	"encoding/json"
	"fmt"
	"io"
	"os"
	"strconv"
	"time"
)

var hostname string

func init() {
	var err error
	if hostname, err = os.Hostname(); err != nil {
		hostname = "<unknown-hostname>"
	}
}

type formatterPlain struct{}

func (f *formatterPlain) emitTimeStamp(w io.Writer) {
	w.Write([]byte(time.Now().Format(time.RFC3339) + " " + hostname + " "))
}

func (f *formatterPlain) formatFileEntry(w io.Writer, file afero.File, description, message string, extra ...string) {
	f.emitTimeStamp(w)
	fmt.Fprintf(w, "%s: %s: %s\n", description, file.Name(), message)
}

func (f *formatterPlain) formatMessage(w io.Writer, format string, a ...interface{}) {
	f.emitTimeStamp(w)
	if format[len(format)-1] != '\n' {
		format += "\n"
	}
	fmt.Fprintf(w, format, a...)
}

func (formatterPlain) finish(w io.Writer) {}

type formatterTSJSON struct {
	initialized bool
}

func (f *formatterTSJSON) emitRecord(w io.Writer, kv ...string) {
	if f.initialized {
		w.Write([]byte(",\n"))
	} else {
		w.Write([]byte("[\n"))
		f.initialized = true
	}
	now := time.Now()
	r := make(map[string]string)
	r["timestamp"] = strconv.Itoa(int(now.UnixNano() / 1000))
	r["datetime"] = now.Format(time.RFC3339)
	r["hostname"] = hostname
	for it := kv; len(it) >= 2; it = it[2:] {
		r[it[0]] = it[1]
	}
	json.NewEncoder(w).Encode(r)
}

func (f *formatterTSJSON) formatFileEntry(w io.Writer, file afero.File, description, message string, extra ...string) {
	fileinfo := []string{"filename", file.Name()}
	if fi, err := file.Stat(); err == nil {
		fileinfo = append(fileinfo, "file_size", strconv.Itoa(int(fi.Size())))
	}
	extra = append([]string{"timestamp_desc", description, "message", message}, extra...)
	extra = append(fileinfo, extra...)
	f.emitRecord(w, extra...)
}

func (f *formatterTSJSON) formatMessage(w io.Writer, format string, a ...interface{}) {
	extra := []string{"timestamp_desc", "msg", "message", fmt.Sprintf(format, a...)}
	f.emitRecord(w, extra...)
}

func (f *formatterTSJSON) finish(w io.Writer) {
	if !f.initialized {
		w.Write([]byte("["))
	}
	w.Write([]byte("]\n"))
}