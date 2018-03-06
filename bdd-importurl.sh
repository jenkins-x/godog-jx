#!/usr/bin/env bash
echo "Running the BDD import URL tests"
go test -test.v -godog.feature=importurl.feature
