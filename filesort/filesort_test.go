package main

import (
	"bufio"
	"bytes"
	"io"
	"io/ioutil"
	"os"
	"strings"
	"testing"
)

func strSliceEquals(a, b []string) bool {
	if len(a) != len(b) {
		return false
	}
	for i, a_str := range a {
		b_str := b[i]
		if a_str != b_str {
			return false
		}
	}
	return true
}

func strSliceSliceEquals(a, b [][]string) bool {
	if len(a) != len(b) {
		return false
	}
	for i, a_group := range a {
		b_group := b[i]
		if !strSliceEquals(a_group, b_group) {
			return false
		}
	}
	return true
}

func TestStrSliceSplit(t *testing.T) {
	var tests = []struct {
		s     []string
		limit int
		want  [][]string
	}{
		{[]string{"1", "2", "3"}, 1, [][]string{
			[]string{"1"},
			[]string{"2"},
			[]string{"3"},
		}},
		{[]string{"1", "2", "3"}, 2, [][]string{
			[]string{"1", "2"},
			[]string{"3"},
		}},
		{[]string{"1", "2", "3"}, 3, [][]string{
			[]string{"1", "2", "3"},
		}},
		{[]string{}, 3, [][]string{}},
	}

	for _, test := range tests {
		result := strSliceSplit(test.s, test.limit)
		if !strSliceSliceEquals(result, test.want) {
			t.Errorf("strSliceSplit(%q,%v) produces: %q, want: %q",
				test.s, test.limit, result, test.want)
		}
	}
}

func buffer(lines []string) *bytes.Buffer {
	var buf bytes.Buffer
	for _, line := range lines {
		buf.WriteString(line)
		buf.WriteByte('\n')
	}
	return &buf
}

func bufferConcat(rs_def [][]string) *bytes.Buffer {
	var b bytes.Buffer
	for _, lines := range rs_def {
		buffer(lines).WriteTo(&b)
	}
	return &b
}

func readers(rs_def [][]string) []io.Reader {
	var rs []io.Reader
	for _, lines := range rs_def {
		rs = append(rs, buffer(lines))
	}
	return rs
}

func TestSourceSet(t *testing.T) {
	var rs_def = [][]string{
		[]string{"b", "d"},
		[]string{"a", "c"},
	}
	ss, err := newSourceSet(readers(rs_def))
	if err != nil {
		t.Errorf("newSourceSet() produces error: %v", err)
	}
	min, err := ss.popMin()
	if err != nil {
		t.Errorf("ss.popMin() produces error: %v", err)
	}
	if min != "a\n" {
		t.Errorf("ss.popMin() produces: %q, want: %q", min, "a\n")
	}
}

var merge_tests = []struct {
	rs   [][]string
	want []string
}{
	{
		[][]string{
			[]string{},
		},
		[]string{},
	},
	{
		[][]string{
			[]string{"a"},
		},
		[]string{"a"},
	},
	{
		[][]string{
			[]string{"a", "c"},
		},
		[]string{"a", "c"},
	},
	{
		[][]string{
			[]string{"a", "c", "d", "e"},
			[]string{"b", "f", "g", "h"},
		},
		[]string{"a", "b", "c", "d", "e", "f", "g", "h"},
	},
	{
		[][]string{
			[]string{"a"},
			[]string{"c"},
			[]string{"d"},
			[]string{"e"},
			[]string{"b"},
			[]string{"f"},
			[]string{"g"},
			[]string{"h"},
			[]string{"j"},
			[]string{"k"},
			[]string{"i"},
		},
		[]string{"a", "b", "c", "d", "e", "f", "g", "h", "i", "j", "k"},
	},
}
var merge_limits = []int{2, 3, 10}

func TestMergeSimple(t *testing.T) {
	const ms = "mergeSimple(rs=%q"

	for _, test := range merge_tests {
		var out bytes.Buffer
		err := mergeSimple(readers(test.rs), &out)
		if err != nil {
			t.Errorf(ms+" produces error: %v", test.rs, err)
		}

		out_str := out.String()
		want_str := buffer(test.want).String()
		if out_str != want_str {
			t.Errorf(ms+" produces: %q, want: %q", test.rs, out_str, want_str)
		}
	}
}

func tempFile(lines []string) string {
	f, err := ioutil.TempFile("", "filesort_test_")
	if err != nil {
		panic("Can't create temp file")
	}
	buf := bufio.NewWriter(f)
	for _, line := range lines {
		_, err := buf.WriteString(line)
		if err == nil {
			err = buf.WriteByte('\n')
		}
		if err != nil {
			panic("Write error")
		}
	}
	err = buf.Flush()
	if err == nil {
		err = f.Close()
	}
	if err != nil {
		panic("Write error")
	}

	return f.Name()
}

func tempFiles(rs_def [][]string) []string {
	var result []string
	for _, lines := range rs_def {
		result = append(result, tempFile(lines))
	}
	return result
}

func TestMerge(t *testing.T) {
	const m = "merge(%q,%v)"

	for _, test := range merge_tests {
		for _, limit := range merge_limits {
			names := tempFiles(test.rs)
			result, err := merge(names, limit)
			if err != nil {
				t.Errorf(m+" produces error: %v", test.rs, limit, err)
			}

			f, err := os.Open(result)
			if err != nil {
				t.Errorf(m+": can't open result file, error: %v",
					test.rs, limit, err)
			}

			result_b, err := ioutil.ReadAll(f)
			if err != nil {
				t.Errorf(m+": can't read result file, error: %v",
					test.rs, limit, err)
			}
			f.Close()
			deleteFile(result)

			result_s := string(result_b)
			want_s := buffer(test.want).String()
			if result_s != want_s {
				t.Errorf(m+" produces: %q, want: %q",
					test.rs, limit, result_s, want_s)
			}
		}
	}
}

func checkReadLines(t *testing.T, in string, limit int, want []string, eof bool) {
	rl := "readLines(in=%q,limit=%v)"
	var r = bufio.NewReader(strings.NewReader(in))

	lines, err := readLines(r, limit)
	if eof && (err != io.EOF) {
		t.Errorf(rl + " does not reach EOF")
	}
	if !eof && (err != nil) {
		t.Errorf(rl+" produces error: %v", in, limit, err)
	}
	if !strSliceEquals(lines, want) {
		t.Errorf(rl+" produces: %v, want: %v",
			in, limit, lines, want)
	}
}

func TestReadLines(t *testing.T) {
	var tests = []struct {
		in    string
		limit int
		out   []string
		eof   bool
	}{
		{"ala\nbala\n", 1, []string{"ala\n"}, false},
		{"ala\nbala\n", 2, []string{"ala\n", "bala\n"}, false},
		{"ala\nbala", 2, []string{"ala\n", "bala\n"}, true},
	}

	for _, test := range tests {
		checkReadLines(t, test.in, test.limit, test.out, test.eof)
	}
}

func TestSplit(t *testing.T) {
	type subcase struct {
		limit int
		want  [][]string
	}
	var tests = []struct {
		in       []string
		subcases []subcase
	}{
		{
			[]string{},
			[]subcase{
				subcase{
					limit: 3,
					want: [][]string{
						[]string{},
					},
				},
				subcase{
					limit: 10,
					want: [][]string{
						[]string{},
					},
				},
			},
		},
		{
			[]string{"a"},
			[]subcase{
				subcase{
					limit: 2,
					want: [][]string{
						[]string{"a"},
					},
				},
				subcase{
					limit: 10,
					want: [][]string{
						[]string{"a"},
					},
				},
			},
		},
		{
			[]string{"a", "c"},
			[]subcase{
				subcase{
					limit: 10,
					want: [][]string{
						[]string{"a", "c"},
					},
				},
			},
		},
		{
			[]string{"a", "b", "c", "d", "e", "f", "g", "h"},
			[]subcase{
				subcase{
					limit: 2,
					want: [][]string{
						[]string{"a", "b"},
						[]string{"c", "d"},
						[]string{"e", "f"},
						[]string{"g", "h"},
					},
				},
				subcase{
					limit: 4,
					want: [][]string{
						[]string{"a", "b", "c", "d"},
						[]string{"e", "f", "g", "h"},
					},
				},
				subcase{
					limit: 7,
					want: [][]string{
						[]string{"a", "b", "c", "d", "e", "f", "g"},
						[]string{"h"},
					},
				},
				subcase{
					limit: 8,
					want: [][]string{
						[]string{"a", "b", "c", "d", "e", "f", "g", "h"},
					},
				},
			},
		},
	}

	for _, test := range tests {
		in := buffer(test.in)
		in_s := in.String()
		for _, sub := range test.subcases {
			names, err := split(in, sub.limit)
			if err != nil {
				t.Errorf("split(%q,%v) produces error: %v",
					in_s, sub.limit, err)
			}
			// FIXME: compare output
			deleteFiles(names)
		}
	}

}

func TestSortLinesWrite(t *testing.T) {
	for _, test := range merge_tests {
		for _, limit := range merge_limits {
			in := bufferConcat(test.rs)
			in_s := in.String()
			var out bytes.Buffer
			err := sortLinesWrite(in, limit, &out)
			if err != nil {
				t.Errorf("sortLinesWrite(%q,%v) produces error: %v",
					in_s, limit, err)
			}

			result_s := out.String()
			want_s := buffer(test.want).String()
			if result_s != want_s {
				t.Errorf("sortLinesWrire(%q,%v) produces: %q, want: %q",
					test.rs, limit, result_s, want_s)
			}
		}
	}
}
