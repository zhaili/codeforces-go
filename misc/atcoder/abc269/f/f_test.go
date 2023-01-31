// Code generated by copypasta/template/atcoder/generator_test.go
package main

import (
	"github.com/EndlessCheng/codeforces-go/main/testutil"
	"testing"
)

// 提交地址：https://atcoder.jp/contests/abc269/submit?taskScreenName=abc269_f
func Test_run(t *testing.T) {
	t.Log("Current test is [f]")
	testCases := [][2]string{
		{
			`5 4
6
1 3 2 4
1 5 1 1
5 5 1 4
4 4 2 2
5 5 4 4
1 5 1 4`,
			`28
27
36
14
0
104`,
		},
		{
			`1000000000 1000000000
3
1000000000 1000000000 1000000000 1000000000
165997482 306594988 719483261 992306147
1 1000000000 1 1000000000`,
			`716070898
240994972
536839100`,
		},
		{
			`999999999 999999999
3
999999999 999999999 999999999 999999999
216499784 840031647 84657913 415448790
1 999999999 1 999999999`,
			`712559605
648737448
540261130`,
		},
		
	}
	testutil.AssertEqualStringCase(t, testCases, 0, run)
}
// https://atcoder.jp/contests/abc269/tasks/abc269_f
