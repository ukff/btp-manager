package main

import (
	"bytes"
	"fmt"
	"os/exec"
	"sort"
	"strings"
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
	var res string
	res += "| No. | CR state    | Condition type | Condition status  | Condition reason | Remark "
	res += "\n"
	res += "|-|-|-|-|-|-|"
	res += "\n"
	no := 1
	for _, row := range rows {
		res += fmt.Sprintf("|%d| %s| %s| %t| %s| %s", no, row.crState, row.conditionType, row.conditionStatus, row.conditionReason, row.remark)
		res += "\n"
		no++
	}
	return res
}

func ss(x []tableRow) {
	sort.Slice(x, func(i, j int) bool {
		return x[i].conditionReason < x[j].conditionReason
	})
}

func main() {

	cmd := exec.Command("/bin/sh", "table.sh")
	var outb, errb bytes.Buffer
	cmd.Stdout = &outb
	cmd.Stderr = &errb
	err := cmd.Run()
	if err != nil {
		fmt.Println(errb.String())
		panic(err)
	}
	p := strings.Split(outb.String(), "====")
	fmt.Println(p[1])
	holder := holder{}
	metadata := strings.Split(p[1], "!")
	for _, v := range metadata {
		fmt.Println(v)
		holder.lineToRow(v)
	}

	ss(holder.ready)
	ss(holder.proccessing)
	ss(holder.deleting)
	ss(holder.error)

	ctr := make([]tableRow, 0)
	ctr = append(ctr, holder.ready...)
	ctr = append(ctr, holder.proccessing...)
	ctr = append(ctr, holder.deleting...)
	ctr = append(ctr, holder.error...)

	expectedTable := renderTable(ctr)
	fmt.Println(expectedTable)
}

func (h *holder) lineToRow(line string) {
	parts := strings.Split(line, "//")
	var state, remark string
	words := strings.Fields(parts[0])
	if len(words) > 0 {
		reason := words[0]
		conditionType := "Ready"
		crState := ""
		if len(parts) >= 2 {
			comments := strings.Split(parts[1], ";")

			if len(comments) > 0 {
				state = comments[0]
				remark = comments[1]
			} else {
				state = "tba"
				remark = "Tba"
			}
		} else {
			state = "tba"
			remark = "Tba"
		}

		conditionStats := calculateConditionStatus(crState, conditionType)
		cleanString(&state)
		cleanString(&crState)
		cleanString(&conditionType)
		cleanString(&remark)
		cleanString(&reason)

		tr := &tableRow{
			crState:         state,
			conditionType:   conditionType,
			conditionStatus: conditionStats,
			conditionReason: reason,
			remark:          remark,
		}

		switch state {
		case "Ready":
			h.ready = append(h.ready, *tr)
		case "Processing":
			h.proccessing = append(h.proccessing, *tr)
		case "Error":
			h.error = append(h.error, *tr)
		case "Deleting":
			h.deleting = append(h.deleting, *tr)
		}
	}
}

func calculateConditionStatus(crState, conditionType string) bool {
	return crState == "Ready" && conditionType == "Ready"
}

func cleanString(s *string) {
	*s = strings.Replace(*s, ":", "", -1)
	*s = strings.Replace(*s, "/", "", -1)
	*s = strings.Replace(*s, ",", "", -1)
}
