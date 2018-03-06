# BDD tests for Jenkins X using Godog

This project contains the BDD tests for using Jenkins X with the [jx command line](https://github.com/jenkins-x/jx) 

## Prerequisits

- __golang__ https://golang.org/doc/install#install
- a Jenkins X installation


## JX BDD tests

To run the `jx import url` tests:

    ./bdd-importurl.sh
    
The bdd tests will use your local jx setup in `~/.jx` and defaults to the current git provider in `~/.jx/gitAuth.yaml`

To specify a specific git provider use:

    export GIT_PROVIDER_URL="github.com"

Passing in the Git provider URL of your choice

### Interactive mode

If you have not setup API tokens for your Jenkins or git provider use interactive mode to run a test:

    export JX_INTERACTIVE="true"

You can then enter the required API tokens and whatnot on the first run. Future tests will not need interactive mode
