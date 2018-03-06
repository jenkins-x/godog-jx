Feature: jx import url
  As a developer
  I need to be able to import an existing git repository and have jx setup the CI / CD

  Scenario: SpringBoot sample application can be imported and built
    Given a temporary directory
    When running 'jx import --url' with "https://github.com/jenkins-x/test-spring-boot-app" in a directory
    Then there should be a jenkins project created
    And the application should be built and promoted via CI / CD
