package leetcode

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/levigross/grequests"
	"github.com/skratchdot/open-golang/open"
)

const cookiePath = "~/.lc/cookie.txt"
const templateCpp = "~/code/cf/lc/lc.cxx"
const templateH = "~/code/cf/lc/lc.hpp"

// use cookie login
func login_by_cookie(cookie string) (session *grequests.Session, err error) {
	const ua = "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/106.0.0.0 Safari/537.36"
	ro := &grequests.RequestOptions{UserAgent: ua,
		Headers: map[string]string{"cookie": cookie}}
	session = grequests.NewSession(ro)
	return
}

// 获取题目信息（含题目链接）
// contestTag 如 "weekly-contest-200"，可以从比赛链接中获取
func GenLeetCodeTestsBySession(session *grequests.Session, contestTag string, openWebPage bool, contestDir, customComment string) error {
	var problems []*problem
	var err error
	for {
		problems, err = fetchProblemURLs(session, contestTag)
		if err == nil {
			break
		}
		fmt.Println(err)
		time.Sleep(500 * time.Millisecond)
	}

	if len(problems) == 0 {
		return nil
	}

	customComment = updateComment(customComment)

	for _, p := range problems {
		p.openURL = openWebPage
		p.customComment = customComment
		p.contestDir = contestDir + "/" + p.id
	}

	fmt.Println("题目链接获取成功，开始解析")
	return handleContestProblems(session, problems)
}

func (p *problem) createContestDir() error {
	return os.MkdirAll(p.contestDir, os.ModePerm)
}

func genCppContent(defaultCode string) []byte {
	const tag = "// @lc code=start\n"
	const methodTag = "#define METHOD"

	content, _ := ioutil.ReadFile(ExpandHomePath(templateCpp))
	pos1 := bytes.Index(content, []byte(tag)) + len(tag)
	pos2 := bytes.Index(content, []byte(methodTag)) + len(methodTag)

	pattern := regexp.MustCompile(" ([_A-z0-9]+)\\(")
	rs := pattern.FindStringSubmatch(defaultCode)
	methodName := rs[1]

	codes := []byte{}
	codes = append(codes, content[:pos1]...)
	codes = append(codes, []byte(defaultCode)...)
	codes = append(codes, content[pos1:pos2]...)
	codes = append(codes, []byte(" "+methodName)...)
	codes = append(codes, content[pos2:]...)

	return codes
}

func (p *problem) writeCppFile() error {
	p.defaultCode = strings.TrimSpace(p.defaultCode)
	fileContent := genCppContent(p.defaultCode)

	filePath := p.contestDir + fmt.Sprintf("/%s", filepath.Base(templateH))
	contentH, _ := ioutil.ReadFile(ExpandHomePath(templateH))
	os.WriteFile(filePath, contentH, 0644)

	filePath = p.contestDir + fmt.Sprintf("/%s.cpp", p.id)
	return os.WriteFile(filePath, fileContent, 0644)
}

func (p *problem) writeTestCaseFile() error {
	lines := []string{strconv.Itoa(len(p.sampleIns)), p.sampleInTexts}
	testDataStr := strings.Join(lines, "\n")

	filePath := p.contestDir + fmt.Sprintf("/in1.txt")
	os.WriteFile(filePath, []byte(testDataStr), 0644)

	lines = []string{}
	for _, outArgs := range p.sampleOuts {
		lines = append(lines, outArgs...)
	}
	testDataStr = strings.Join(lines, "\n")

	filePath = p.contestDir + fmt.Sprintf("/ans1.txt")
	return os.WriteFile(filePath, []byte(testDataStr), 0644)
}

func handleContestProblems(session *grequests.Session, problems []*problem) error {
	wg := &sync.WaitGroup{}
	wg.Add(1 + len(problems))

	go func() {
		defer wg.Done()
		for _, p := range problems {
			if p.openURL {
				if err := open.Run(p.url); err != nil {
					fmt.Println("open err:", p.url, err)
				}
			}
		}
	}()

	for _, p := range problems {
		fmt.Println(p.id, p.url)

		go func(p *problem) {
			defer wg.Done()

			if err := p.parseHTML(session, "cpp"); err != nil {
				fmt.Fprintln(os.Stderr, err)
			}

			if err := p.createContestDir(); err != nil {
				fmt.Println("createDir err:", p.url, err)
				return
			}

			if err := p.writeCppFile(); err != nil {
				fmt.Println("writeFile err:", p.url, err)
			}

			if err := p.writeTestCaseFile(); err != nil {
				fmt.Println("writeTestFile err:", p.url, err)
			}

			p.writeTestDataFile()

		}(p)
	}

	wg.Wait()
	return nil
}

func ExpandHomePath(path string) string {
	if strings.HasPrefix(path, "~/") {
		dirname, _ := os.UserHomeDir()
		path = filepath.Join(dirname, path[2:])
	}
	return path
}

func FetchContestData(contestid string) {
	content, err := ioutil.ReadFile(ExpandHomePath(cookiePath))
	if err != nil {
		log.Fatal(err)
		return
	}
	cookie_str := strings.TrimSuffix(string(content), "\n")

	ses, err := login_by_cookie(cookie_str)
	if err != nil {
		fmt.Println(err)
		return
	}
	tag := fmt.Sprintf("weekly-contest-%s", contestid)
	GenLeetCodeTestsBySession(ses, tag, false, tag, "")
}
