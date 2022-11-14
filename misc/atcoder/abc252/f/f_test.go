// Code generated by copypasta/template/atcoder/generator_test.go
package main

import (
	"github.com/EndlessCheng/codeforces-go/main/testutil"
	"testing"
)

// 提交地址：https://atcoder.jp/contests/abc252/submit?taskScreenName=abc252_f
func Test_run(t *testing.T) {
	t.Log("Current test is [f]")
	testCases := [][2]string{
		{
			`5 7
1 2 1 2 1`,
			`16`,
		},
		{
			`3 1000000000000000
1000000000 1000000000 1000000000`,
			`1000005000000000`,
		},
		
	}
	testutil.AssertEqualStringCase(t, testCases, 0, run)
}
// https://atcoder.jp/contests/abc252/tasks/abc252_f
