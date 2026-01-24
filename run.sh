#!/bin/bash

go build -o grip_bin ./cmd/grip-web/main.go && ./grip_bin
