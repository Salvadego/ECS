//go:build ignore

package main

import (
	"bytes"
	"flag"
	"fmt"
	"go/format"
	"os"
	"text/template"
)

var (
	tmplFile = flag.String("template", "queries.tmpl", "path to template")
	outFile  = flag.String("out", "queries_gen.go", "output file")
	maxArity = flag.Int("max", 5, "maximum query arity")
)

func main() {
	flag.Parse()

	data := struct {
		ArityList []int
		MaxArity  int
	}{
		ArityList: make([]int, *maxArity),
		MaxArity:  *maxArity,
	}
	for i := 1; i <= *maxArity; i++ {
		data.ArityList[i-1] = i
	}

	// template funcs
	funcMap := template.FuncMap{
		"seq": func(n int) []int {
			s := make([]int, n)
			for i := range s {
				s[i] = i + 1
			}
			return s
		},
		"genTypeParams": func(n int) []string {
			s := make([]string, n)
			for i := range s {
				s[i] = fmt.Sprintf("C%d", i+1)
			}
			return s
		},
	}

	tmplBytes, err := os.ReadFile(*tmplFile)
	if err != nil {
		panic(err)
	}
	tmpl, err := template.New("queries").Funcs(funcMap).Parse(string(tmplBytes))
	if err != nil {
		panic(err)
	}

	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, data); err != nil {
		panic(err)
	}

	// gofmt
	src, err := format.Source(buf.Bytes())
	if err != nil {
		fmt.Println("Warning: could not format code:", err)
		src = buf.Bytes()
	}

	if err := os.WriteFile(*outFile, src, 0644); err != nil {
		panic(err)
	}
}
