package deepequal

import (
	"bytes"
	"reflect"
	"regexp"
	"strings"
)

// TestPrinter basically a masque for testing.XXX.
type TestPrinter interface {
	Log(a ...any)
	Error(a ...any)
}

// SideBySide outputs a and b side by side with a difference highlight.
func SideBySide[T any](p TestPrinter, what string, want, got T) {
	lv := reflect.ValueOf(want)
	rv := reflect.ValueOf(got)

	if !Equal(want, got) {
		p.Error("mismatched expected and actual values of", what)
	} else {
		p.Log(`match for expected and actual values of`, what)
	}
	printDiff(p, lv, rv)

}

func printDiff(p TestPrinter, l, r reflect.Value) {
	diff := difference(l, r, false)

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
	rrs := append([]string{formatBold + "Actual\u001B[0m"}, rdrs...)

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
