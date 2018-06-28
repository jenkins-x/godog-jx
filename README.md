# BDD tests for Jenkins X using Godog

This project contains the BDD tests for using Jenkins X with the [jx command line](https://github.com/jenkins-x/jx) 

## Prerequisits

- __golang__ https://golang.org/doc/install#install
- a Jenkins X installation


## Running the BDD tests

To specify a specific git provider and the git user/organisation to use for the tests:

    export GIT_PROVIDER_URL="github.com"
    export GIT_ORGANISATION="jstrachan"

The bdd tests will use your local jx setup in `~/.jx` and defaults to the current git provider in `~/.jx/gitAuth.yaml`

If you wish to use a different location for `~/.jx` due to running on a CI system then use:

    export JX_HOME=/foo/bar

To setup the git auth token first before running the tests you may wanna use a command like this:

	  jx create git token -n MyGitServerName MyGitUser -t MyToken
    
Then to run the `jx spring` tests:

    make jx-spring
    
### Interactive mode

If you have not setup API tokens for your Jenkins or git provider use interactive mode to run a test:

    export JX_INTERACTIVE="true"

You can then enter the required API tokens and whatnot on the first run. Future tests will not need interactive mode
