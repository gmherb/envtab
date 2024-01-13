#!/usr/bin/env bash

set -euxo pipefail

pkgs=$(go list ./... 2> /dev/null | grep -v /vendor/)
deps=`echo ${pkgs} | tr ' ' ","`
echo "mode: atomic" > profile.tmp

for pkg in $pkgs; do
    go test -v -race -cover -coverpkg "$deps" -coverprofile=profile.tmp $pkg

    if [ -f profile.tmp ]; then
        tail -n +2 profile.tmp >> profile.cov
        rm profile.tmp
    fi
done;

go tool cover -func=profile.cov