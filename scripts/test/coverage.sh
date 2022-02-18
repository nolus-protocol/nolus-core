#!/bin/bash
#
# Code coverage generation


# Display the global code coverage
go tool cover -func=coverage.txt ;

# If needed, generate HTML report
if [ "$1" == "html" ]; then
    go tool cover -html=coverage.txt -o coverage.html ;
fi

