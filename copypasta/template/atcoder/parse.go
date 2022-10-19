package atcoder

import (
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"strings"
	"sync"

	"github.com/levigross/grequests"
)

const cookiePath = "~/.atc/cookie.txt"
const templateCpp = "~/code/cf/template.cpp"

func expandHomePath(path string) string {
	if strings.HasPrefix(path, "~/") {
		dirname, _ := os.UserHomeDir()
		path = filepath.Join(dirname, path[2:])
	}
	return path
}

func createDir(path string) error {
	return os.MkdirAll(path, os.ModePerm)
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

func writeCppFile(contestDir, id string) error {
	fileContent, _ := ioutil.ReadFile(expandHomePath(templateCpp))
	filePath := fmt.Sprintf("%s/%s.cpp", contestDir, id)
	return os.WriteFile(filePath, fileContent, 0644)
}

func writeTestCaseFile(contestDir string, sampleIns, sampleOuts []string) {
	for i, sample := range sampleIns {
		filePath := fmt.Sprintf("%s/in%d.txt", contestDir, i+1)
		os.WriteFile(filePath, []byte(sample), 0644)
	}

	for i, sample := range sampleOuts {
		filePath := fmt.Sprintf("%s/ans%d.txt", contestDir, i+1)
		os.WriteFile(filePath, []byte(sample), 0644)
	}
}

func handleProblem(session *grequests.Session, contestID string, id string) (err error) {
	problemURL := fmt.Sprintf("https://atcoder.jp/contests/%[1]s/tasks/%[1]s_%[2]s", contestID, id)
	sampleIns, sampleOuts, err := parseTask(session, problemURL)

	contestDir := fmt.Sprintf("%s/%s", contestID, id)
	createDir(contestDir)

	writeTestCaseFile(contestDir, sampleIns, sampleOuts)
	writeCppFile(contestDir, id)

	return err
}

func handleContest(session *grequests.Session, contestID string) (err error) {
	taskNum, err := fetchTaskNum(contestID)
	if err != nil {
		return err
	}
	fmt.Printf("共 %d 道题目\n", taskNum)

	fmt.Println("开始解析样例输入输出")
	wg := &sync.WaitGroup{}
	defer wg.Wait()
	for i := 0; i < taskNum; i++ {
		wg.Add(1)
		// we don't want spent too much time on waiting responses one by one, so we use goroutine!
		go func(id string) {
			defer wg.Done()
			defer func() {
				if err := recover(); err != nil {
					fmt.Println("[error]", string(id), err)
				}
			}()
			if err := handleProblem(session, contestID, id); err != nil {
				fmt.Println("[error]", string(id), err)
				return
			}
		}(string('a' + byte(i)))
	}

	return nil
}

func FetchProblem(taskID string) (err error) {
	session, err := loginByCookie(cookiePath)

	c := taskID[len(taskID)-1:][0]
	if (c >= 'a' && c <= 'z') || (c >= 'A' && c <= 'Z') {
		contestID := taskID[:len(taskID)-1]
		id := taskID[len(taskID)-1:]
		return handleProblem(session, contestID, id)
	} else {
		return handleContest(session, taskID)
	}
}
