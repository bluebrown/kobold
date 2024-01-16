package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/bluebrown/kobold/kioutil"
	"sigs.k8s.io/kustomize/kyaml/kio"
)

func main() {

	var (
		uri      string
		r        kio.Reader
		w        kio.Writer
		setAnnos bool
		base     string
	)

	flag.BoolVar(&setAnnos, "a", false, "set path annotation")
	flag.StringVar(&base, "b", "", "base branch, if not provided is the same as the destination branch")
	flag.Parse()

	if len(flag.Args()) < 2 {
		fmt.Fprintf(os.Stderr, "not enough arguments: %d\n", len(os.Args))
		os.Exit(1)
	}

	cmd := flag.Arg(0)
	uri = flag.Arg(1)

	if cmd == "source" {
		rt := kioutil.NewGitPackageReader(uri)
		rt.SetPathAnnotation = setAnnos
		r = rt
		w = &kio.ByteWriter{Writer: os.Stdout}

	} else if cmd == "sink" {
		r = &kio.ByteReader{Reader: os.Stdin}
		w = kioutil.NewGitPackageWriter(uri, base)
	} else {
		fmt.Fprintf(os.Stderr, "invalid command: %s\n", cmd)
		os.Exit(1)
	}

	if err := kioutil.CopyIO(r, w); err != nil {
		fmt.Fprintf(os.Stderr, "failed to copy nodes: %v\n", err)
		os.Exit(1)
	}
}
