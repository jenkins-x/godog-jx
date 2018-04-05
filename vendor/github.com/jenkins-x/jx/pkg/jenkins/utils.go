package jenkins

import (
	"crypto/tls"
	"errors"
	"fmt"
	"net/http"

	"io"
	"os"

	"github.com/jenkins-x/golang-jenkins"
	jenkauth "github.com/jenkins-x/jx/pkg/auth"
	"github.com/jenkins-x/jx/pkg/util"
)

func GetJenkinsClient(url string, batch bool, configService *jenkauth.AuthConfigService) (*gojenkins.Jenkins, error) {
	if url == "" {
		return nil, errors.New("no Jenkins service be found in the development namespace!\nAre you sure you installed Jenkins X? Try: http://jenkins-x.io/getting-started/")
	}
	tokenUrl := JenkinsTokenURL(url)

	auth := jenkauth.CreateAuthUserFromEnvironment("JENKINS")
	username := auth.Username
	var err error
	config := configService.Config()

	showForm := false
	if auth.IsInvalid() {
		// lets try load the current auth
		config, err = configService.LoadConfig()
		if err != nil {
			return nil, err
		}
		auths := config.FindUserAuths(url)
		if len(auths) > 1 {
			// TODO choose an auth
		}
		showForm = true
		a := config.FindUserAuth(url, username)
		if a != nil {
			if a.IsInvalid() {
				auth, err = EditUserAuth(url, configService, config, a, tokenUrl, batch)
				if err != nil {
					return nil, err
				}
			} else {
				auth = *a
			}
		} else {
			// lets create a new Auth
			auth, err = EditUserAuth(url, configService, config, &auth, tokenUrl, batch)
			if err != nil {
				return nil, err
			}
		}
	}

	if auth.IsInvalid() {
		if showForm {
			return nil, fmt.Errorf("No valid Username and API Token specified for Jenkins server: %s\n", url)
		} else {
			fmt.Printf("No $JENKINS_USERNAME and $JENKINS_TOKEN environment variables defined!\n")
			PrintGetTokenFromURL(os.Stdout, tokenUrl)
			if batch {
				fmt.Printf("Then run this command on your terminal and try again:\n\n")
				fmt.Printf("export JENKINS_TOKEN=myApiToken\n\n")
				return nil, errors.New("No environment variables (JENKINS_USERNAME and JENKINS_TOKEN) or JENKINS_BEARER_TOKEN defined")
			}
		}
	}

	jauth := &gojenkins.Auth{
		Username:    auth.Username,
		ApiToken:    auth.ApiToken,
		BearerToken: auth.BearerToken,
	}
	jenkins := gojenkins.NewJenkins(jauth, url)

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

func PrintGetTokenFromURL(out io.Writer, tokenUrl string) (int, error) {
	return fmt.Fprintf(out, "Please go to %s and click %s to get your API Token\n", util.ColorInfo(tokenUrl), util.ColorInfo("Show API Token"))
}

func JenkinsTokenURL(url string) string {
	tokenUrl := util.UrlJoin(url, "/me/configure")
	return tokenUrl
}

func EditUserAuth(url string, configService *jenkauth.AuthConfigService, config *jenkauth.AuthConfig, auth *jenkauth.UserAuth, tokenUrl string, batchMode bool) (jenkauth.UserAuth, error) {

	fmt.Printf("\nTo be able to connect to the Jenkins server we need a username and API Token\n\n")

	f := func(username string) error {
		fmt.Printf("\nPlease go to %s and click %s to get your API Token\n", util.ColorInfo(tokenUrl), util.ColorInfo("Show API Token"))
		fmt.Printf("Then COPY the API token so that you can paste it into the form below:\n\n")
		return nil
	}

	defaultUsername := "admin"

	err := config.EditUserAuth("Jenkins", auth, defaultUsername, true, batchMode, f)
	if err != nil {
		return *auth, err
	}
	err = configService.SaveUserAuth(url, auth)
	return *auth, err
}
