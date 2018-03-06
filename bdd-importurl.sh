#!/usr/bin/env bash
export GO15VENDOREXPERIMENT="1"
echo "Running the BDD import URL tests"
go test -test.v -godog.feature=importurl.feature
