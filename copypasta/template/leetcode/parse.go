package leetcode

import (
	"bytes"
	"encoding/json"
	"fmt"
	"html"
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

func expandHomePath(path string) string {
	if strings.HasPrefix(path, "~/") {
		dirname, _ := os.UserHomeDir()
		path = filepath.Join(dirname, path[2:])
	}
	return path
}

// use cookie login
func loginByCookie(cookiePath string) (session *grequests.Session, err error) {
	const ua = "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/106.0.0.0 Safari/537.36"
	content, err := ioutil.ReadFile(expandHomePath(cookiePath))
	if err != nil {
		log.Fatal(err)
		return
	}
	cookie := strings.TrimSuffix(string(content), "\n")
	ro := &grequests.RequestOptions{UserAgent: ua,
		Headers: map[string]string{"cookie": cookie}}
	session = grequests.NewSession(ro)
	return
}

func queryProblemSlug(session *grequests.Session, problemid string) (titleSlug string) {
	ro := &grequests.RequestOptions{
		JSON: map[string]interface{}{
			"variables": map[string]interface{}{
				"categorySlug": "",
				"skip":         0,
				"limit":        50,
				"filters":      map[string]string{"searchKeywords": problemid},
			},
			"query": "\n    query problemsetQuestionList($categorySlug: String, $limit: Int, $skip: Int, $filters: QuestionListFilterInput) {\n  problemsetQuestionList(\n    categorySlug: $categorySlug\n    limit: $limit\n    skip: $skip\n    filters: $filters\n  ) {\n    hasMore\n    total\n    questions {\n      acRate\n      difficulty\n      freqBar\n      frontendQuestionId\n      isFavor\n      paidOnly\n      solutionNum\n      status\n      title\n      titleCn\n      titleSlug\n      topicTags {\n        name\n        nameTranslated\n        id\n        slug\n      }\n      extra {\n        hasVideoSolution\n        topCompanyTags {\n          imgUrl\n          slug\n          numSubscribed\n        }\n      }\n    }\n  }\n}\n    ",
		}}
	resp, err := session.Post(graphqlURL, ro)
	var data map[string]interface{}
	if err = resp.JSON(&data); err != nil {
		fmt.Printf("could not unmarshal json: %s\n", err)
		return
	}

	for _, question := range data["data"].(map[string]interface{})["problemsetQuestionList"].(map[string]interface{})["questions"].([]interface{}) {
		q := question.(map[string]interface{})
		if q["frontendQuestionId"] == problemid {
			titleSlug = q["titleSlug"].(string)
			break
		}
	}
	return titleSlug
}

func queryProblemDetail(session *grequests.Session, titleSlug string) (sampleOuts [][]string, sampleIns string, codes string) {
	ro := &grequests.RequestOptions{
		JSON: map[string]interface{}{
			"operationName": "getQuestionDetail",
			"variables":     map[string]string{"titleSlug": titleSlug},
			"query": `
query getQuestionDetail($titleSlug: String!) {
  question(titleSlug: $titleSlug) {
    content
    stats
    codeDefinition
    sampleTestCase
    exampleTestcases
    enableRunCode
    metaData
    translatedContent
  }
}`}}

	resp, err := session.Post(graphqlURL, ro)
	if err != nil {
		fmt.Println("could get problem detail")
		return
	}

	var data map[string]interface{}
	if err = resp.JSON(&data); err != nil {
		fmt.Printf("could not unmarshal json: %s\n", err)
		return
	}

	question := data["data"].(map[string]interface{})["question"].(map[string]interface{})

	content := question["content"].(string)

	for {
		start := strings.Index(content, "<pre>")
		end := strings.Index(content, "</pre>")
		if start == -1 || end == -1 {
			break
		}

		testCase := content[start : end+6]
		sampleOuts = append(sampleOuts, []string{parseCase(testCase)})

		content = content[end+6:]
	}
	sampleIns = question["exampleTestcases"].(string)
	codes = parseCodeDefinition(question["codeDefinition"].(string), "cpp")
	return
}

func parseCodeDefinition(jsonText string, lang string) string {
	data := []struct {
		Value       string `json:"value"`
		DefaultCode string `json:"defaultCode"`
	}{}
	if err := json.Unmarshal([]byte(jsonText), &data); err != nil {
		return ""
	}

	for _, template := range data {
		if template.Value == lang {
			return strings.TrimSpace(template.DefaultCode)
		}
	}
	return ""
}

func parseCase(text string) string {
	patternStart := regexp.MustCompile("utput:?</strong>")
	pos := patternStart.FindStringIndex(text)
	if pos == nil {
		return ""
	}
	start := pos[1]

	patternEnd := regexp.MustCompile("\n.*Explanation")
	posEnd := patternEnd.FindStringIndex(text)
	var end = -1
	if posEnd == nil {
		end = strings.LastIndex(text, "</pre>")
	} else {
		end = posEnd[0]
	}
	text = strings.TrimSpace(text[start:end])
	text = html.UnescapeString(text)
	return text
}

func FetchProblem(problemid string) (err error) {
	session, err := loginByCookie(cookiePath)

	titleSlug := queryProblemSlug(session, problemid)
	if len(titleSlug) == 0 {
		fmt.Printf("Problem %s not found.\n", problemid)
		return
	}

	sampleOuts, sampleIns, codes := queryProblemDetail(session, titleSlug)

	var p problem
	p.id = titleSlug

	p.defaultCode = codes
	p.sampleInTexts = sampleIns
	p.sampleOuts = sampleOuts
	p.contestDir = "p" + problemid

	if err = p.createContestDir(); err != nil {
		fmt.Println("createDir err:", p.url, err)
		return
	}

	if err = p.writeCppFile(); err != nil {
		fmt.Println("writeFile err:", p.url, err)
	}

	if err = p.writeTestCaseFile(); err != nil {
		fmt.Println("writeTestFile err:", p.url, err)
	}

	p.writeTestDataFile()

	return err
}

func (p *problem) createContestDir() error {
	return os.MkdirAll(p.contestDir, os.ModePerm)
}

func genCppContent(defaultCode string) []byte {
	const tag = "// @lc code=start\n"
	const methodTag = "#define METHOD"

	content, _ := ioutil.ReadFile(expandHomePath(templateCpp))
	pos1 := bytes.Index(content, []byte(tag)) + len(tag)
	pos2 := bytes.Index(content, []byte(methodTag)) + len(methodTag)

	solutionStart := strings.Index(defaultCode, "class Solution")
	pattern := regexp.MustCompile(" ([_A-z0-9]+)\\(")
	rs := pattern.FindStringSubmatch(defaultCode[solutionStart:])
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
	contentH, _ := ioutil.ReadFile(expandHomePath(templateH))
	os.WriteFile(filePath, contentH, 0644)

	filePath = p.contestDir + fmt.Sprintf("/%s.cpp", p.id)
	return os.WriteFile(filePath, fileContent, 0644)
}

func (p *problem) writeTestCaseFile() error {
	lines := []string{strconv.Itoa(len(p.sampleOuts)), p.sampleInTexts}
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

func FetchContestData(contestid string) {
	ses, err := loginByCookie(cookiePath)
	if err != nil {
		fmt.Println(err)
		return
	}
	var tag string
	if contestid[0] == 'b' {
		tag = fmt.Sprintf("biweekly-contest-%s", contestid[1:])
	} else {
		tag = fmt.Sprintf("weekly-contest-%s", contestid)
	}
	GenLeetCodeTestsBySession(ses, tag, false, tag, "")
}
