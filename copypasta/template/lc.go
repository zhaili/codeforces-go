package main

import (
	lc "github.com/EndlessCheng/codeforces-go/copypasta/template/leetcode"
	docopt "github.com/docopt/docopt-go"
)

func main() {
	var usage = `Usage:
  lc race <contestid>
  lc parse <problemid>
  lc -h | --help | --version`

	opts, _ := docopt.ParseDoc(usage)
	if value, _ := opts.Bool("race"); value {
		contestid, _ := opts.String("<contestid>")
		lc.FetchContestData(contestid)
	} else if value, _ := opts.Bool("parse"); value {
		problemid, _ := opts.String("<problemid>")
		lc.FetchProblem(problemid)
	}
}
