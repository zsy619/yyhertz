package version

import (
	"io"
	"log"
	"os"
	"runtime"
	"text/template"
	"time"
)

// RuntimeInfo holds information about the current runtime.
type RuntimeInfo struct {
	GoVersion string
	GOOS      string
	GOARCH    string
	NumCPU    int
	GOPATH    string
	GOROOT    string
	Compiler  string
	Version   string
	Published string
}

// InitBanner loads the banner and prints it to output
// All errors are ignored, the application will not
// print the banner in case of error.
func InitBanner(out io.Writer, in io.Reader) {
	if in == nil {
		log.Fatal("The input is nil")
	}

	banner, err := io.ReadAll(in)
	if err != nil {
		log.Fatalf("Error while trying to read the banner: %s", err)
	}

	show(out, string(banner))
}

func show(out io.Writer, content string) {
	t, err := template.New("banner").
		Funcs(template.FuncMap{"Now": Now}).
		Parse(content)

	if err != nil {
		log.Fatalf("Cannot parse the banner template: %s", err)
	}

	err = t.Execute(out, RuntimeInfo{
		runtime.Version(),
		runtime.GOOS,
		runtime.GOARCH,
		runtime.NumCPU(),
		os.Getenv("GOPATH"),
		runtime.GOROOT(),
		runtime.Compiler,
		Version,
		time.Now().Format("2006-01-02 15:04:05"),
	})
	if err != nil {
		log.Fatalf("Error while trying to execute the banner template: %s", err)
	}
}

// Now returns the current local time in the specified layout
func Now(layout string) string {
	return time.Now().Format(layout)
}

type outputMode int

const (
	_ outputMode = iota
	DiscardNonColorEscSeq
	OutputNonColorEscSeq
)

type colorWriter struct {
	w    io.Writer
	mode outputMode
}

func (cw *colorWriter) Write(p []byte) (int, error) {
	return cw.w.Write(p)
}

func NewColorWriter(w io.Writer) io.Writer {
	return NewModeColorWriter(w, DiscardNonColorEscSeq)
}

// NewModeColorWriter create and initializes a new ansiColorWriter
// by specifying the outputMode.
func NewModeColorWriter(w io.Writer, mode outputMode) io.Writer {
	if _, ok := w.(*colorWriter); !ok {
		return &colorWriter{
			w:    w,
			mode: mode,
		}
	}
	return w
}
