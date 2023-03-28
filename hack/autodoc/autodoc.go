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
	spaceMargin             = 10
	errorExitCode           = 1
	okExitCode              = 0
	expectedDataChunksCount = 3
	dataSeparator           = "==="
)

type tableRow struct {
	groupOrder      int
	crState         string
	conditionType   string
	conditionStatus bool
	conditionReason string
	remark          string
}

func main() {
	dataForProcessing := getConditionsData()
	dataChunks := strings.Split(dataForProcessing, "====")
	if len(dataChunks) != expectedDataChunksCount {
		fmt.Println("'extract_conditions_data.sh' data output failed, it should contain 3 elements")
		os.Exit(errorExitCode)
	}

	constReasons := getConstReasons(dataChunks[0])
	errors, reasonsMetadata := getAndValidateReasonsMetadata(dataChunks[1])
	if len(errors) > 0 {
		printErrors(errors)
		os.Exit(errorExitCode)
	}

	errors = checkIfConstsAndMetadataAreInSync(constReasons, reasonsMetadata)
	if len(errors) > 0 {
		fmt.Println("The declared reasons in const Go section are out out sync with Reasons metadata")
		printErrors(errors)
		os.Exit(errorExitCode)
	}

	tableFromMdFile := tableToStruct(dataChunks[2])
	errors = compareContent(tableFromMdFile, reasonsMetadata)
	if len(errors) > 0 {
		printErrors(errors)
		fmt.Println("Below can be found auto-generated table which contain new changes:")
		fmt.Println(renderTable(reasonsMetadata))
		os.Exit(errorExitCode)
	}

	os.Exit(okExitCode)
}

func getConditionsData() string {
	cmd := exec.Command("/bin/sh", "hack/autodoc/extract_conditions_data.sh")
	var cmdOut, cmdErr bytes.Buffer
	cmd.Stdout = &cmdOut
	cmd.Stderr = &cmdErr
	if err := cmd.Run(); err != nil {
		fmt.Errorf(cmdErr.String())
		os.Exit(errorExitCode)
	}
	return cmdOut.String()
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

func getAndValidateReasonsMetadata(input string) ([]string, []tableRow) {
	reasonMetadata := strings.Split(input, "\n")
	allTableRows := make([]tableRow, 0)
	errors := make([]string, 0)
	for _, dataLine := range reasonMetadata {
		err, lineStructured := tryConvertGoLineToStruct(dataLine)
		if err != nil {
			errors = append(errors, err.Error())
			continue
		}
		if lineStructured != nil {
			allTableRows = append(allTableRows, *lineStructured)
		}
	}

	return errors, allTableRows
}

func checkIfConstsAndMetadataAreInSync(constReasons []string, reasonsMetadata []tableRow) []string {
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
			errors = append(errors, fmt.Sprintf("there is a Reason = (%s) declarated in const scope, but there is no matching metadata for it", constReason))
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

func compareContent(currentTableStructured []tableRow, newTableStructured []tableRow) []string {
	errors := make([]string, 0)

	logValidationFailedMessage := func(a, b, s string) string {
		if a != b {
			return fmt.Sprintf("Docs are not synced with Go code, difference detected in reason (%s), current value in docs is (%s) but newer in Go code is (%s)", s, a, b)
		}
		return ""
	}

	for _, newRow := range newTableStructured {
		found := false
		for _, currentRow := range currentTableStructured {
			if newRow.conditionReason == currentRow.conditionReason {
				found = true
				if newRow.remark != currentRow.remark {
					errors = append(errors, logValidationFailedMessage(currentRow.remark, newRow.remark, newRow.conditionReason))
					break
				}

				if newRow.conditionStatus != currentRow.conditionStatus {
					errors = append(errors, logValidationFailedMessage(strconv.FormatBool(currentRow.conditionStatus), strconv.FormatBool(newRow.conditionStatus), newRow.conditionReason))
					break
				}

				if newRow.crState != currentRow.crState {
					errors = append(errors, logValidationFailedMessage(currentRow.crState, newRow.crState, newRow.conditionReason))
					break
				}

				if newRow.conditionType != currentRow.conditionType {
					errors = append(errors, logValidationFailedMessage(currentRow.conditionType, newRow.conditionType, newRow.conditionReason))
					break
				}
				break
			}
		}

		if !found {
			errors = append(errors, fmt.Sprintf("Reason (%s) not found in docs.", newRow.conditionReason))
		}
	}
	return errors
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

func tryConvertGoLineToStruct(line string) (error, *tableRow) {
	if line == "" {
		return fmt.Errorf("empty line given"), nil
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

	calculateConditionStatus := func(state, conditionType string) bool {
		return state == "Ready" && conditionType == "Ready"
	}

	return nil, &tableRow{
		groupOrder:      detectGroupOrder(state),
		crState:         state,
		conditionType:   conditionType,
		conditionStatus: calculateConditionStatus(state, conditionType),
		conditionReason: reason,
		remark:          remark,
	}
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
		return 6
	}
}

func printErrors(errors []string) {
	for _, error := range errors {
		fmt.Println(error)
	}
}

func cleanString(s *string) {
	*s = strings.Replace(*s, " ", "", -1)
	*s = strings.Replace(*s, ":", "", -1)
	*s = strings.Replace(*s, "/", "", -1)
	*s = strings.Replace(*s, ",", "", -1)
}
