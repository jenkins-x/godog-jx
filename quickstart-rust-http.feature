Feature: jx create quickstart rust-http
  As a developer
  I need to be able to create and delete a new rust-http application from a quickstart with CI / CD

  Scenario: Create and delete a new rust-http quickstart application
    Given a work directory
    When running "jx create quickstart -b -f rust-http" in that directory
    Then there should be a jenkins project created
    And the application should be built and promoted via CI / CD
    And the application should be deleted after running jx delete app
    And the git repo should be deleted after running jx delete repo
