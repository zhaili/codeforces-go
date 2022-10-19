package main

import (
	atc "github.com/EndlessCheng/codeforces-go/copypasta/template/atcoder"
	docopt "github.com/docopt/docopt-go"
)

func main() {
	var usage = `Usage:
  atc parse <taskid>
  atc -h | --help | --version`

	opts, _ := docopt.ParseDoc(usage)
	if value, _ := opts.Bool("parse"); value {
		taskid, _ := opts.String("<taskid>")
		atc.FetchProblem(taskid)
	}
}
