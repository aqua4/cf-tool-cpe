package client

import (
	"errors"
	"fmt"
	"net/url"
	"regexp"
	"strings"

	"cf-tool/util"
)

func findMessage(body []byte) (string, error) {
	reg := regexp.MustCompile(`Codeforces.showMessage\("([^"]*)"\);\s*?Codeforces\.reformatTimes\(\);`)
	tmp := reg.FindSubmatch(body)
	if tmp != nil {
		return string(tmp[1]), nil
	}
	return "", errors.New("Cannot find any message")
}

func findErrorMessage(body []byte) (string, error) {
	reg := regexp.MustCompile(`error[a-zA-Z_\-\ ]*">(.*?)</span>`)
	tmp := reg.FindSubmatch(body)
	if tmp == nil {
		return "", errors.New("Cannot find error")
	}
	return string(tmp[1]), nil
}

// Submit submit (block while pending)
func (c *Client) Submit(info Info, langID, source string) (err error) {
	fmt.Printf("Submit %v using %v\n", info.Hint(), Langs[langID])

	URL, err := info.SubmitURL(c.host)
	if err != nil {
		return
	}

	body, err := util.GetBody(c.client, URL)
	if err != nil {
		return
	}

	handle, err := findHandle(body)
	if err != nil {
		return
	}

	fmt.Printf("Current user: %v\n", handle)

	csrf, err := findCsrf(body)
	if err != nil {
		return
	}

	body, err = util.PostBody(c.client, fmt.Sprintf("%v?csrf_token=%v", URL, csrf), url.Values{
		"csrf_token":            {csrf},
		"ftaa":                  {c.Ftaa},
		"bfaa":                  {c.Bfaa},
		"action":                {"submitSolutionFormSubmitted"},
		"submittedProblemIndex": {info.ProblemID},
		"programTypeId":         {langID},
		"contestId":             {info.ContestID},
		"source":                {source},
		"tabSize":               {"4"},
		"_tta":                  {"594"},
		"sourceCodeConfirmed":   {"true"},
	})
	if err != nil {
		return
	}

	errMsg, err := findErrorMessage(body)
	if err == nil {
		return errors.New(errMsg)
	}

	msg, err := findMessage(body)
	if err != nil {
		return errors.New("Submit failed")
	}
	if !strings.Contains(msg, "submitted successfully") {
		return errors.New(msg)
	}

	fmt.Println("Submitted")

	if err = c.WatchSubmission(info); err != nil {
		return
	}

	c.Handle = handle
	return c.save()
}
