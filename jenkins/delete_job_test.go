package jenkins

import (
	"fmt"
	"strings"

	"github.com/DATA-DOG/godog"
	"github.com/jenkins-x/godog-jx/utils"
)

func thereIsAJobCalled(jobExpression string) error {
	jobName := utils.ReplaceEnvVars(jobExpression)
	jenkins, err := utils.GetJenkinsClient()
	if err != nil {
		return fmt.Errorf("error getting a Jenkins client %v", err)
	}
	paths := strings.Split(jobName, "/")
	_, err = jenkins.GetJobByPath(paths...)
	if err != nil {
		return fmt.Errorf("error finding existing job %s due to %s", jobName, err)
	}
	return nil
}

func iDeleteTheJob(jobExpression string) error {
	jobName := utils.ReplaceEnvVars(jobExpression)
	jenkins, err := utils.GetJenkinsClient()
	if err != nil {
		return fmt.Errorf("error getting a Jenkins client  %v", err)
	}
	paths := strings.Split(jobName, "/")
	job, err := jenkins.GetJobByPath(paths...)
	if err != nil {
		return fmt.Errorf("error finding existing job %s due to %s", jobName, err)
	}
	err = jenkins.DeleteJob(job)
	if err != nil {
		return fmt.Errorf("error deleteing job %s due to %s", job.Name, err)
	}
	return nil
}

func thereShouldNotBeAJob(jobExpression string) error {
	jobName := utils.ReplaceEnvVars(jobExpression)
	jenkins, err := utils.GetJenkinsClient()
	if err != nil {
		return fmt.Errorf("error getting a Jenkins client  %v", err)
	}

	job, err := jenkins.GetJob(jobName)
	if err != nil {
		return nil
	}
	return fmt.Errorf("error found existing job %s", job.Name)
}

func DeleteJobFeatureContext(s *godog.Suite) {
	s.Step(`^there is a job called "([^"]*)"$`, thereIsAJobCalled)
	s.Step(`^I delete the "([^"]*)" job$`, iDeleteTheJob)
	s.Step(`^there should not be a "([^"]*)" job$`, thereShouldNotBeAJob)
}
