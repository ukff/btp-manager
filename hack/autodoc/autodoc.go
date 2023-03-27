package main

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"sort"
	"strconv"
	"strings"
)

const (
	spaceMargin int = 10
	staticLineLen
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
	renderElement := func(x int, s string, c string) string {
		x = x - len(s)
		e := ""
		e += s
		for i := 0; i < x+spaceMargin; i++ {
			e += c
		}
		return e
	}

	sort.Slice(rows, func(i, j int) bool {
		if rows[i].groupOrder != rows[j].groupOrder {
			return rows[i].groupOrder < rows[j].groupOrder
		}
		return rows[i].conditionReason < rows[j].conditionReason
	})

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
	mdTable += fmt.Sprintf("| %s | %s | %s | %s | %s | %s |", renderElement(10, "No.", " "), renderElement(10, "CR state", " "), renderElement(10, "Condition type", " "), renderElement(10, "Condition status", " "), renderElement(longestConditionReasons, "Condition reason", " "), renderElement(longestRemark, "Remark", " "))
	mdTable += "\n"
	mdTable += fmt.Sprintf("| %s | %s | %s | %s | %s | %s |", renderElement(10, "", "-"), renderElement(10, "", "-"), renderElement(10, "", "-"), renderElement(10, "", "-"), renderElement(longestConditionReasons, "", "-"), renderElement(longestRemark, "", "-"))
	mdTable += "\n"

	lineNumber := 1
	for _, row := range rows {
		mdTable += fmt.Sprintf("| %s | %s | %s | %s | %s | %s |", renderElement(10, strconv.Itoa(lineNumber), " "), renderElement(10, row.crState, " "), renderElement(10, row.conditionType, " "), renderElement(10, strconv.FormatBool(row.conditionStatus), " "), renderElement(longestConditionReasons, row.conditionReason, " "), renderElement(longestRemark, row.remark, " "))
		mdTable += "\n"
		lineNumber++
	}

	return mdTable
}

func getConstReasons(input string) []string {
	result := make([]string, 0)
	words := strings.Split(input, "\n")
	for _, v := range words {
		words2 := strings.Fields(v)
		if len(words2) > 0 {
			result = append(result, words2[0])
		}
	}

	return result
}

func getReasonsMetadata(input string) ([]error, []tableRow) {
	reasonMetadata := strings.Split(input, "\n")
	allTableRows := make([]tableRow, 0)
	errors := make([]error, 0)
	for _, dataLine := range reasonMetadata {
		err, l := stringToStruct(dataLine)
		if err != nil {
			errors = append(errors, err)
		}
		if l != nil {
			allTableRows = append(allTableRows, *l)
		}
	}

	return errors, allTableRows
}

func checkIfReasonsAndConstReasonsAreSynced(constReasons []string, reasonsMetadata []tableRow) []string {
	errors := make([]string, 0)
	checkIfConstReasonHaveMetadata := func(constReason string) bool {
		for _, reasonMetadata := range reasonsMetadata {
			if reasonMetadata.conditionReason == constReason {
				return true
			}
		}
		return false
	}

	for _, constReason := range constReasons {
		if !checkIfConstReasonHaveMetadata(constReason) {
			err := fmt.Sprintf("there is a Reason = (%s) declarated in const scope, but there is no matching metadata for it", constReason)
			errors = append(errors, err)
		}
	}
	return errors
}

func tableToStruct(tableMd string) []tableRow {
	rows := strings.Split(tableMd, "\n")
	rows = rows[1 : len(rows)-2]
	trableStructured := make([]tableRow, 0)
	for i, s := range rows {
		if i == 0 || i == 1 || i == 2 {
			continue
		}
		cleanLine := strings.Split(s, "|")
		cleanLine = cleanLine[1 : len(cleanLine)-1]
		crState := cleanLine[1]
		cleanString(&crState)
		conditionType := cleanLine[2]
		cleanString(&conditionType)
		conditionStatus, _ := strconv.ParseBool(cleanLine[3])
		conditionReason := cleanLine[4]
		cleanString(&conditionReason)
		remark := cleanLine[5]
		cleanString(&remark)

		tr := tableRow{
			groupOrder:      detectGroupOrder(crState),
			crState:         crState,
			conditionType:   conditionType,
			conditionStatus: conditionStatus,
			conditionReason: conditionReason,
			remark:          remark,
		}
		trableStructured = append(trableStructured, tr)
	}
	return trableStructured
}

func main() {
	rawData := getRawData()
	dataParts := strings.Split(rawData, "====")
	if len(dataParts) != 3 {
		fmt.Println("sh returned wrong data")
		os.Exit(1)
	}

	constReasons := getConstReasons(dataParts[0])
	errs, reasonsMetadata := getReasonsMetadata(dataParts[1])
	if len(errs) > 0 {
		for _, v := range errs {
			fmt.Println(v)
		}
		os.Exit(1)
	}

	errors := checkIfReasonsAndConstReasonsAreSynced(constReasons, reasonsMetadata)
	if len(errors) > 0 {
		fmt.Println("The declarated reasons in consts are out out sync with reasons metadata. Please sync it.")
		for _, error := range errors {
			fmt.Println(error)
		}
	}

	tableFromMdFile := tableToStruct(dataParts[2])
	compareContent(tableFromMdFile, reasonsMetadata)
}

func compareContent(currentTableStructured []tableRow, newTableStructured []tableRow) {
	errors := make([]string, 0)
	okContent := true
	for _, ed := range newTableStructured {
		ok := false
		for _, cts := range currentTableStructured {
			if ed.conditionReason == cts.conditionReason {
				if ed.remark != ed.remark {
					ok = false
					break
				}

				if ed.conditionStatus == ed.conditionStatus {
					ok = false
					break
				}

				if ed.crState == ed.crState {
					ok = false
					break
				}

				if ed.conditionType == ed.conditionType {
					ok = false
					break
				}
				ok = true
				break
			}
		}

		if !ok {
			okContent = false
			err := fmt.Sprintf("new reason: %s not found in documentation")
			errors = append(errors, err)
		}
	}

	if !okContent {
		fmt.Println(renderTable(newTableStructured))
	}
}

func stringToStruct(line string) (error, *tableRow) {
	if line == "" {
		return nil, nil
	}
	line = strings.Replace(line, "\n", "", -1)
	parts := strings.Split(line, "//")
	if len(parts) != 2 {
		return fmt.Errorf("validation failed! -> in line (%s) there is no comment section (//) included, comment section should have following format (//CRState;Remark)", line), nil
	}

	words := strings.Fields(parts[0])
	if len(words) != 2 {
		return fmt.Errorf("validation failed! -> line (%s) is bad structured, it should have following format (Reason: TypeAndStatus, //CRState;Remark", line), nil
	}

	comments := strings.Split(parts[1], ";")
	if len(comments) != 2 {
		return fmt.Errorf("validation failed! -> comment in line (%s) is bad structured, it should have following format (//CRState;Remark)", line), nil
	}

	reason := words[0]
	cleanString(&reason)

	conditionType := "Ready"
	cleanString(&conditionType)

	state := comments[0]
	cleanString(&state)

	remark := comments[1]
	cleanString(&remark)

	return nil, &tableRow{
		groupOrder:      detectGroupOrder(state),
		crState:         state,
		conditionType:   conditionType,
		conditionStatus: calculateConditionStatus(state, conditionType),
		conditionReason: reason,
		remark:          remark,
	}
}

func getRawData() string {
	cmdd := exec.Command("ls")
	var cmddOut, cmddErr bytes.Buffer
	cmdd.Stdout = &cmddOut
	cmdd.Stderr = &cmddErr
	if err := cmdd.Run(); err != nil {
		fmt.Println(cmddErr.String())
		panic(err)
	}

	cmd := exec.Command("/bin/sh", "hack/autodoc/extract_conditions_data.sh")
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
	fmt.Println(state)
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
		return 6
	}
}

func calculateConditionStatus(state, conditionType string) bool {
	return state == "Ready" && conditionType == "Ready"
}

func cleanString(s *string) {
	*s = strings.Replace(*s, " ", "", -1)
	*s = strings.Replace(*s, ":", "", -1)
	*s = strings.Replace(*s, "/", "", -1)
	*s = strings.Replace(*s, ",", "", -1)
}
