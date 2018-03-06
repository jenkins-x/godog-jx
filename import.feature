Feature: jx import
  As a developer
  I need to be able to import an existing directory and have jx setup the CI / CD

  Scenario: SpringBoot sample application can be imported and built
    Given a directory containing a Spring Boot application
    When running "jx import" in that directory
    Then there should be a jenkins project created
    And the application should be built and promoted via CI / CD
