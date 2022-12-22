// Code generated by copypasta/template/atcoder/generator_test.go
package main

import (
	"github.com/EndlessCheng/codeforces-go/main/testutil"
	"testing"
)

// 提交地址：https://atcoder.jp/contests/keyence2020/submit?taskScreenName=keyence2020_d
func Test_run(t *testing.T) {
	t.Log("Current test is [d]")
	testCases := [][2]string{
		{
			`3
3 4 3
3 2 3`,
			`1`,
		},
		{
			`2
2 1
1 2`,
			`-1`,
		},
		{
			`4
1 2 3 4
5 6 7 8`,
			`0`,
		},
		{
			`5
28 15 22 43 31
20 22 43 33 32`,
			`-1`,
		},
		{
			`5
4 46 6 38 43
33 15 18 27 37`,
			`3`,
		},
		
	}
	testutil.AssertEqualStringCase(t, testCases, 0, run)
}
// https://atcoder.jp/contests/keyence2020/tasks/keyence2020_d
