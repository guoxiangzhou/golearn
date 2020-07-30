#!/bin/bash
rm -rf a.log
go run main.go 2>&1 | tee err.log
