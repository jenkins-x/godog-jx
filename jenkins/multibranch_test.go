package jenkins

import (
	"fmt"
	"os"
	"strings"

	"github.com/DATA-DOG/godog"
	"github.com/jenkins-x/godog-jx/kubernetes"
	"github.com/jenkins-x/godog-jx/utils"
)

func thereIsAJenkinsCredential(credentialID string) error {
	jenkins, err := utils.GetJenkinsClient()
	if err != nil {
		return fmt.Errorf("error getting a Jenkins client %v", err)
	}

	githubUser := os.Getenv("GITHUB_USER")
	if githubUser == "" {
		return fmt.Errorf("GITHUB_USER env var not set")
	}

	githubPassword := os.Getenv("GITHUB_PASSWORD")
	if githubUser == "" {
		return fmt.Errorf("GITHUB_PASSWORD env var not set")
	}

	err = jenkins.CreateCredential(credentialID, githubUser, githubPassword)
	if err != nil {
		return fmt.Errorf("error creating jenkins credential %s %v", "bdd-test", err)
	}
	return nil
}

func weCreateAMultibranchJobCalled(jobName string) error {
	jenkins, err := utils.GetJenkinsClient()
	if err != nil {
		return fmt.Errorf("error getting a Jenkins client %v", err)
	}
	jobXML, err := utils.GetFileAsString("resources/multi_job.xml")
	if err != nil {
		return err
	}

	githubUser := os.Getenv("GITHUB_USER")
	if githubUser == "" {
		return fmt.Errorf("GITHUB_USER env var not set")
	}

	jobXML = strings.Replace(jobXML, "$GITHUB_USER", githubUser, 1)
	err = jenkins.CreateJobWithXML(jobXML, jobName)
	if err != nil {
		return fmt.Errorf("error creating Job %v", err)
	}
	return nil
}

func triggerAScanOfTheJob(jobName string) error {

	jenkins, err := utils.GetJenkinsClient()
	if err != nil {
		return fmt.Errorf("error getting a Jenkins client %v", err)
	}

	job, err := jenkins.GetJob(jobName)
	if err != nil {
		return fmt.Errorf("error creating Job %v", err)
	}

	err = jenkins.Build(job, nil)
	if err != nil {
		return fmt.Errorf("error triggering job %s %v", jobName, err)
	}
	return nil

}

func thereShouldBeAJobThatCompletesSuccessfully(jobName string) error {
	jenkins, err := utils.GetJenkinsClient()
	if err != nil {
		return fmt.Errorf("error getting a Jenkins client %v", err)
	}
	return ThereShouldBeAJobThatCompletesSuccessfully(jobName, jenkins)
}

func theApplicationIsInTheEnvironment(app, state, environment string) error {
	return kubernetes.CheckPodStatus(app, state, environment)
}

func MultibranchFeatureContext(s *godog.Suite) {
	s.Step(`^there is a "([^"]*)" jenkins credential$`, thereIsAJenkinsCredential)
	s.Step(`^we create a multibranch job called "([^"]*)"$`, weCreateAMultibranchJobCalled)
	s.Step(`^trigger a scan of the job "([^"]*)"$`, triggerAScanOfTheJob)
	s.Step(`^there should be a "([^"]*)" job that completes successfully$`, thereShouldBeAJobThatCompletesSuccessfully)
	s.Step(`^the "([^"]*)" application is "([^"]*)" in the "([^"]*)" environment$`, theApplicationIsInTheEnvironment)
}
