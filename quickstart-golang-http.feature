Feature: jx create quickstart golang-http
  As a developer
  I need to be able to create and delete a new golang-http application from a quickstart with CI / CD

  Scenario: Create and delete a new golang-http quickstart application
    Given a work directory
    When running "jx create quickstart -b -f golang-http" in that directory
    Then there should be a jenkins project created
    And the application should be built and promoted via CI / CD
    And the application should be deleted after running jx delete app
    And the git repo should be deleted after running jx delete repo
