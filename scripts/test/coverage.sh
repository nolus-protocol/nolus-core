#!/bin/bash
#
# Code coverage generation


# Display the global code coverage
go tool cover -func=coverage.txt ;

# If needed, generate XML report
if [ "$1" = "xml" ]
then
    go install github.com/boumenot/gocover-cobertura
    gocover-cobertura < cover.out > coverage.xml
fi

# If needed, generate HTML report
if [ "$1" = "html" ]
then
    go tool cover -html=cover.out -o coverage.html ;
fi
