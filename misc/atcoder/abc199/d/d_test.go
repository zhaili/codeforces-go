// Code generated by copypasta/template/atcoder/generator_test.go
package main

import (
	"github.com/EndlessCheng/codeforces-go/main/testutil"
	"testing"
)

// 提交地址：https://atcoder.jp/contests/abc199/submit?taskScreenName=abc199_d
func Test_run(t *testing.T) {
	t.Log("Current test is [d]")
	testCases := [][2]string{
		{
			`3 3
1 2
2 3
3 1`,
			`6`,
		},
		{
			`3 0`,
			`27`,
		},
		{
			`4 6
1 2
2 3
3 4
2 4
1 3
1 4`,
			`0`,
		},
		{
			`20 0`,
			`3486784401`,
		},
		// TODO 测试参数的下界和上界
		{
			`20 19
1 2
2 3
3 4
4 5
5 6
6 7
7 8
8 9
9 10
10 11
11 12
12 13
13 14
14 15
15 16
16 17
17 18
18 19
19 20`,
			``,
		},
	}
	testutil.AssertEqualStringCase(t, testCases, 1, run)
}
// https://atcoder.jp/contests/abc199/tasks/abc199_d
