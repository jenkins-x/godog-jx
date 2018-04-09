Feature: fork GitHub Repo
  In order to test importing a quickstart
  As a tester
  I need to be able to fork a Bitbucket repository to a clean fork

  Scenario: Fork repository
    Given there is no fork of "atlassian/python-bitbucket"
    When I fork the "atlassian/python-bitbucket" Bitbucket repo to the current user
    Then there should be a fork for the current user which has the same last commit as "atlassian/python-bitbucket"

