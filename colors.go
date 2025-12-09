package main

import "github.com/fatih/color"

var (
	Success   = color.New(color.FgHiGreen).SprintFunc()
	Info      = color.New(color.FgCyan).SprintFunc()
	Warn      = color.New(color.FgYellow).SprintfFunc()
	Error     = color.New(color.FgRed).SprintFunc()
	Important = color.New(color.FgHiMagenta).SprintfFunc()
)
