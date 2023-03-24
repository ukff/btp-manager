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
	groupOrder      int
	crState         string
	conditionType   string
	conditionStatus bool
	conditionReason string
	remark          string
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

func main() {
	rawData := getRawData()
	dataParts := strings.Split(rawData, "====")
	reasonMetadata := strings.Split(dataParts[1], "!")
	orginalTable := dataParts[2]
	allTableRows := make([]tableRow, 0)
	for _, dataLine := range reasonMetadata {
		l := lineToRow(dataLine)
		if l != nil {
			allTableRows = append(allTableRows, *l)
		}
	}

	sort.Slice(allTableRows, func(i, j int) bool {
		if allTableRows[i].groupOrder != allTableRows[j].groupOrder {
			return allTableRows[i].groupOrder < allTableRows[j].groupOrder
		}
		return allTableRows[i].conditionReason < allTableRows[j].conditionReason
	})

	expectedTable := renderTable(allTableRows)
	fmt.Println(expectedTable)
	fmt.Println(orginalTable)
}

func lineToRow(line string) *tableRow {
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

		cleanString(&state)
		cleanString(&conditionType)
		cleanString(&remark)
		cleanString(&reason)

		return &tableRow{
			groupOrder:      detectGroupOrder(state),
			crState:         state,
			conditionType:   conditionType,
			conditionStatus: calculateConditionStatus(state, conditionType),
			conditionReason: reason,
			remark:          remark,
		}
	}

	return nil
}

func getRawData() string {
	cmd := exec.Command("/bin/sh", "table.sh")
	var cmdOut, cmdErr bytes.Buffer
	cmd.Stdout = &cmdOut
	cmd.Stderr = &cmdErr
	if err := cmd.Run(); err != nil {
		fmt.Println(cmdErr.String())
		panic(err)
	}
	return cmdOut.String()
}

func detectGroupOrder(state string) int {
	switch state {
	case "Ready":
		return 1
	case "Processing":
		return 2
	case "Deleting":
		return 3
	case "Error":
		return 4
	case "NA":
		return 5
	default:
		return 5
	}
}

func calculateConditionStatus(state, conditionType string) bool {
	return state == "Ready" && conditionType == "Ready"
}

func cleanString(s *string) {
	*s = strings.Replace(*s, ":", "", -1)
	*s = strings.Replace(*s, "/", "", -1)
	*s = strings.Replace(*s, ",", "", -1)
}
