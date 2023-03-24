package main

import (
	"bytes"
	"fmt"
	"os/exec"
	"sort"
	"strconv"
	"strings"
)

const (
	spaceMargin int = 10
)

type tableRow struct {
	crState         string
	conditionType   string
	conditionStatus bool
	conditionReason string
	remark          string
}

type holder struct {
	ready       []tableRow
	proccessing []tableRow
	deleting    []tableRow
	error       []tableRow
}

func renderTable(rows []tableRow) string {
	longestConditionReasons := 0
	for _, row := range rows {
		l := len(row.conditionReason)
		if l > longestConditionReasons {
			longestConditionReasons = l
		}
	}

	longestRemark := 0
	for _, row := range rows {
		l := len(row.remark)
		if l > longestRemark {
			longestRemark = l
		}
	}

	var mdTable string
	mdTable += fmt.Sprintf("| %s | %s | %s | %s | %s | %s |", appendEmptySpace(10, "No.", " "), appendEmptySpace(10, "CR state", " "), appendEmptySpace(10, "Condition type", " "), appendEmptySpace(10, "Condition status", " "), appendEmptySpace(longestConditionReasons, "Condition reason", " "), appendEmptySpace(longestRemark, "Remark", " "))
	mdTable += "\n"
	mdTable += fmt.Sprintf("| %s | %s | %s | %s | %s | %s |", appendEmptySpace(10, "", "-"), appendEmptySpace(10, "", "-"), appendEmptySpace(10, "", "-"), appendEmptySpace(10, "", "-"), appendEmptySpace(longestConditionReasons, "", "-"), appendEmptySpace(longestRemark, "", "-"))
	mdTable += "\n"

	lineNumber := 1
	for _, row := range rows {
		mdTable += fmt.Sprintf("| %s | %s | %s | %s | %s | %s |", appendEmptySpace(10, strconv.Itoa(lineNumber), " "), appendEmptySpace(10, row.crState, " "), appendEmptySpace(10, row.conditionType, " "), appendEmptySpace(10, strconv.FormatBool(row.conditionStatus), " "), appendEmptySpace(longestConditionReasons, row.conditionReason, " "), appendEmptySpace(longestRemark, row.remark, " "))
		mdTable += "\n"
		lineNumber++
	}

	return mdTable
}

func appendEmptySpace(x int, s string, c string) string {
	x = x - len(s)
	e := ""
	e += s
	for i := 0; i < x+spaceMargin; i++ {
		e += c
	}
	return e
}

func sortByConditionReasons(x []tableRow) {
	sort.Slice(x, func(i, j int) bool {
		return x[i].conditionReason < x[j].conditionReason
	})
}

func main() {
	cmd := exec.Command("/bin/sh", "table.sh")
	var cmdOut, cmdErr bytes.Buffer
	cmd.Stdout = &cmdOut
	cmd.Stderr = &cmdErr
	if err := cmd.Run(); err != nil {
		fmt.Println(cmdErr.String())
		panic(err)
	}

	inputData := strings.Split(cmdOut.String(), "====")
	holder := holder{}
	reasonMetadata := strings.Split(inputData[1], "!")
	for _, dataLine := range reasonMetadata {
		holder.lineToRow(dataLine)
	}

	sortByConditionReasons(holder.ready)
	sortByConditionReasons(holder.proccessing)
	sortByConditionReasons(holder.deleting)
	sortByConditionReasons(holder.error)

	allTableRows := make([]tableRow, 0)
	allTableRows = append(allTableRows, holder.ready...)
	allTableRows = append(allTableRows, holder.proccessing...)
	allTableRows = append(allTableRows, holder.deleting...)
	allTableRows = append(allTableRows, holder.error...)

	expectedTable := renderTable(allTableRows)
	fmt.Println(expectedTable)
}

func (h *holder) lineToRow(line string) {
	parts := strings.Split(line, "//")
	var state, remark string
	words := strings.Fields(parts[0])
	if len(words) > 0 {
		reason := words[0]
		conditionType := "Ready"
		if len(parts) >= 2 {
			comments := strings.Split(parts[1], ";")
			if len(comments) > 0 {
				state = comments[0]
				remark = comments[1]
			}
		}

		conditionStats := state == "Ready" && conditionType == "Ready"
		cleanString(&state)
		cleanString(&conditionType)
		cleanString(&remark)
		cleanString(&reason)

		row := &tableRow{
			crState:         state,
			conditionType:   conditionType,
			conditionStatus: conditionStats,
			conditionReason: reason,
			remark:          remark,
		}

		switch state {
		case "Ready":
			h.ready = append(h.ready, *row)
		case "Processing":
			h.proccessing = append(h.proccessing, *row)
		case "Error":
			h.error = append(h.error, *row)
		case "Deleting":
			h.deleting = append(h.deleting, *row)
		}
	}
}

func cleanString(s *string) {
	*s = strings.Replace(*s, ":", "", -1)
	*s = strings.Replace(*s, "/", "", -1)
	*s = strings.Replace(*s, ",", "", -1)
}
