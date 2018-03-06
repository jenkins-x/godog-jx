Feature: Jenkins X multibranch pipeline
  In order to provide a basic jenkins-x experience
  As a project admin
  I need to be able to import and run a SpringBoot Github project via the multibranch plugin

  Scenario: a SpringBoot sample application pipeline builds and deploys successfully
    Given there is a "bdd-test" jenkins credential
    When we create a multibranch job called "spring-boot-web-example"
    And trigger a scan of the job "spring-boot-web-example"
    Then there should be a "spring-boot-web-example/master" job that completes successfully
    And the "spring-boot-web-example" application is "running" in the "staging" environment

#  Scenario: Delete application
#    Given there is a job called "spring-boot-web-example"
#    When I delete the "spring-boot-web-example" job
#    Then there should not be a "spring-boot-web-example" job