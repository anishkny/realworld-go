#!/usr/bin/env bash
set -euxo pipefail

# Install go-ignore-cov if not already installed
export GOPATH=$(go env GOPATH)
export GO_IGNORE_COV_BIN=$GOPATH/bin/go-ignore-cov
which $GO_IGNORE_COV_BIN || go install github.com/hexira/go-ignore-cov@latest

# Convert coverage data to text format
go tool covdata textfmt -i coverage/ -o coverage/coverage.txt

# Remove ignored lines from coverage report
$GO_IGNORE_COV_BIN -f coverage/coverage.txt

# Generate coverage report
go tool cover -func=coverage/coverage.txt | tee coverage/func-coverage.txt

# Generate HTML report
go tool cover -html=coverage/coverage.txt -o coverage/report.html
echo Coverage report generated at: `pwd`/coverage/report.html

# Make sure last line has 100% coverage
# total:				(statements)		100.0%
tail -n 1 coverage/func-coverage.txt | grep -q '100.0%' || (echo "FAIL: Coverage is less than 100%" && exit 1) && echo "OK"
