#!/bin/bash

eval $(go env)

mkdir -p ~/.local/bin/ && curl -o ~/.local/bin/docgen https://github.com/overmindtech/docgen/releases/latest/download/docgen-${GOARCH} && chmod +x ~/.local/bin/docgen

go install github.com/cosmtrek/air@latest
