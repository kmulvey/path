# path
[![path](https://github.com/kmulvey/path/actions/workflows/release_build.yml/badge.svg)](https://github.com/kmulvey/path/actions/workflows/release_build.yml) [![codecov](https://codecov.io/gh/kmulvey/path/branch/main/graph/badge.svg?token=uzpd1I3osO)](https://codecov.io/gh/kmulvey/path) [![Go Report Card](https://goreportcard.com/badge/github.com/kmulvey/path)](https://goreportcard.com/report/github.com/kmulvey/path) [![Go Reference](https://pkg.go.dev/badge/github.com/kmulvey/path.svg)](https://pkg.go.dev/github.com/kmulvey/path)

## Overview
A simple library to handle file path input in Go. 

## Features
- Hanlde absolute and relative paths
- Globbing (must be quoted)
- List files in directories recursivly
- Optional regex to filter results
- Cli via [flag](https://pkg.go.dev/flag), see [example](https://github.com/kmulvey/path/blob/main/cmd/main.go)
