package deepequal

import (
	"bytes"
	"fmt"
	"path/filepath"
	"reflect"
	"regexp"
	"runtime"
	"strings"
)

// TestPrinter basically a masque for testing.XXX.
type TestPrinter interface {
	Helper()
	Log(a ...any)
	Error(a ...any)
}

// SideBySide outputs a and b side by side with a difference highlight.
func SideBySide[T any](p TestPrinter, what string, want, got T) {
	lv := reflect.ValueOf(want)
	rv := reflect.ValueOf(got)

	// Look for *_test.go file in the call stack to show proper line.

	p.Helper()
	if !Equal(want, got) {
		p.Error("mismatched expected and actual values of", what)
	} else {
		p.Log(`a match for expected and actual values of`, what)
	}

	printDiff(p, lv, rv)
}

// linePrefix returns the first call position (<file>:<line>) made in some
// *_test.go file. The result is either empty or is ended with the space.
func linePrefix() string {
	fpcs := make([]uintptr, 128)
	frames := runtime.CallersFrames(fpcs)
	n := runtime.Callers(0, fpcs)
	if n == 0 {
		return ""
	}
	fpcs = fpcs[:n]
	frames = runtime.CallersFrames(fpcs)

	for {
		frame, more := frames.Next()
		if strings.HasSuffix(frame.File, "_test.go") {
			_, name := filepath.Split(frame.File)
			return fmt.Sprintf("\r    %s:%d ", name, frame.Line)
		}

		if !more {
			break
		}
	}

	return ""
}

func printDiff(p TestPrinter, l, r reflect.Value) {
	diff := difference(l, r, false, walkSet{})

	lp := newPrinter(true)
	lp.printValue("", l, diff, false, true, map[uintptr]struct{}{})

	rp := newPrinter(false)
	rp.printValue("", r, diff, false, true, map[uintptr]struct{}{})

	ldrs := strings.Split(lp.buf.String(), "\n")
	rdrs := strings.Split(rp.buf.String(), "\n")
	for i := range ldrs {
		ldrs[i] = strings.Trim(ldrs[i], "\r\n")
	}
	for i := range rdrs {
		rdrs[i] = strings.Trim(rdrs[i], "\r\n")
	}

	lrs := append([]string{formatBold + "Expected\033[0m"}, ldrs...)
	var strips []string
	rrs := append([]string{formatBold + "Actual\033[0m"}, rdrs...)

	max1 := 0
	for _, lr := range lrs {
		v := stripANSI(lr)
		strips = append(strips, v)
		ll := len(v)
		if ll > max1 {
			max1 = ll
		}
	}
	max1++

	rs := len(lrs)
	if len(rrs) > rs {
		rs = len(rrs)
	}

	var res bytes.Buffer
	for i := 0; i < rs; i++ {
		var lc string
		var rc string

		if i < len(lrs) {
			lc = lrs[i]
		} else {
			lc = strings.Repeat(" ", max1)
		}
		if i < len(rrs) {
			rc = rrs[i]
		}

		lc = strings.Trim(lc, "\r\n")
		if i < len(lrs) {
			lc += strings.Repeat(" ", max1-len(strips[i]))
		}

		rc = strings.Trim(rc, "\r\n")

		res.WriteString(lc)
		res.WriteByte(' ')
		res.WriteString(rc)
		res.WriteString("\n\r")
	}

	p.Log("\r" + res.String())
}

const ansi = "[\u001B\u009B][[\\]()#;?]*(?:(?:(?:[a-zA-Z\\d]*(?:;[a-zA-Z\\d]*)*)?\u0007)|(?:(?:\\d{1,4}(?:;\\d{0,4})*)?[\\dA-PRZcf-ntqry=><~]))"

var re = regexp.MustCompile(ansi)

func stripANSI(str string) string {
	return re.ReplaceAllString(str, "")
}
