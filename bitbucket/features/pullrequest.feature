Feature: create Bitbucket pull request
	In order to test promoting to environments
	As a tester
	I need to be able to create a Bitbucket pull request

  # We're going to define environment variables for Bitbucket creds, repo name, branches, etc.
	Scenario: Create pull request
		Given I have defined a valid Bitbucket organization, repo, and branches
		When I create a pull request
		Then there should be a pull request in the defined organization and repo