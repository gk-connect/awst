package cmd

import (
	"testing"

	"github.com/fatih/color"
)

func TestPrintTableDemo(t *testing.T) {
	color.NoColor = false
	if visibleWidth(color.RedString("abc")) != 3 {
		t.Fatal("visibleWidth should ignore ANSI codes")
	}
	rows := [][]string{
		{"web-2", "i-0b", "stopped", "t3.small", "10.0.1.6", "-", "-"},
		{"Api", "i-0a", "running", "t3.micro", "10.0.1.5", "13.1.1.1", "Ok"},
		{"db", "i-0c", "running", "r6g.large", "10.0.2.9", "-", "Failed"},
	}
	printTable([]interface{}{"Name", "ID", "State", "Type", "Private IP", "Public IP", "Checks"}, rows, 0, 2, 6)
	footer(3, "instances", dim(" · ")+"2 "+stateColor("running")+dim(" · ")+"1 "+stateColor("stopped"))
	if rows[0][0] != "Api" {
		t.Fatal("rows should be sorted case-insensitively by name")
	}
}
