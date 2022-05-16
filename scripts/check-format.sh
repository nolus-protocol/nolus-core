#!/bin/sh

echo "`find . -type f -name '*.go' | wc -l` file(s) to check"

all=`{ gofmt -l . | sed -e 's/^/__OUT__/g'; } 2>&1`
out=`echo "$all" | grep "^__OUT__" | sed -e 's/^__OUT__//g'`
err=`echo "$all" | grep -v "^__OUT__"`

ret_code=0
if test -n "$err"; then
    printf "Following errors found in Go code:\n\n%s\n\n" "$err"
    ret_code=1
fi

if test -n "$out"; then
    printf "Please re-check format of following Go source files:\n\n%s\n\n" "$out"
    ret_code=1
fi

exit $ret_code
