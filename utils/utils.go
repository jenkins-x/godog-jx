package utils

import (
	"crypto/tls"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/jenkins-x/golang-jenkins"
)

func GetJenkinsClient() (*gojenkins.Jenkins, error) {
	url := os.Getenv("BDD_JENKINS_URL")
	if url == "" {
		return nil, errors.New("no BDD_JENKINS_URL env var set. Try running this command first:\n\n  eval $(gofabric8 bdd-env)\n")
	}
	username := os.Getenv("BDD_JENKINS_USERNAME")
	token := os.Getenv("BDD_JENKINS_TOKEN")

	bearerToken := os.Getenv("BDD_JENKINS_BEARER_TOKEN")
	if bearerToken == "" && (token == "" || username == "") {
		return nil, errors.New("no BDD_JENKINS_TOKEN or BDD_JENKINS_BEARER_TOKEN && BDD_JENKINS_USERNAME env var set")
	}

	auth := &gojenkins.Auth{
		Username:    username,
		ApiToken:    token,
		BearerToken: bearerToken,
	}
	jenkins := gojenkins.NewJenkins(auth, url)

	// handle insecure TLS for minishift
	httpClient := &http.Client{
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
		},
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			return http.ErrUseLastResponse
		}}
	jenkins.SetHTTPClient(httpClient)
	return jenkins, nil
}

func GetFileAsString(path string) (string, error) {
	buf, err := ioutil.ReadFile(path)
	if err != nil {
		return "", fmt.Errorf("No file found at path %s", path)
	}

	return string(buf), nil
}

type MultiError struct {
	Errors []error
}

func RetryAfter(attempts int, callback func() error, d time.Duration) (err error) {
	m := MultiError{}
	for i := 0; i < attempts; i++ {
		err = callback()
		if err == nil {
			return nil
		}
		m.Collect(err)
		time.Sleep(d)
	}
	return m.ToError()
}

func (m *MultiError) Collect(err error) {
	if err != nil {
		m.Errors = append(m.Errors, err)
	}
}

func (m MultiError) ToError() error {
	if len(m.Errors) == 0 {
		return nil
	}

	errStrings := []string{}
	for _, err := range m.Errors {
		errStrings = append(errStrings, err.Error())
	}
	return fmt.Errorf(strings.Join(errStrings, "\n"))
}
