package leetcode

import (
	"encoding/json"
	"fmt"
	"html"
	"regexp"
	"strings"

	"github.com/levigross/grequests"
)

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
