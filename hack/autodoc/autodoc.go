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
	defaultMdElementSize    = 10
	dataSeparator           = "==="
)

type reasonMetadata struct {
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

	tableFromMdFile := mdTableToStruct(dataChunks[2])
	errors = compareContent(tableFromMdFile, reasonsMetadata)
	if len(errors) > 0 {
		printErrors(errors)
		fmt.Println("Below can be found auto-generated table which contain new changes:")
		fmt.Println(buildMdTable(reasonsMetadata))
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

func getAndValidateReasonsMetadata(input string) ([]string, []reasonMetadata) {
	reasonMetadataw := strings.Split(input, "\n")
	allTableRows := make([]reasonMetadata, 0)
	errors := make([]string, 0)
	for _, dataLine := range reasonMetadataw {
		if dataLine == "" {
			continue
		}
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

func checkIfConstsAndMetadataAreInSync(constReasons []string, reasonsMetadata []reasonMetadata) []string {
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

func mdTableToStruct(tableMd string) []reasonMetadata {
	mdRows := strings.Split(tableMd, "\n")
	mdRows = mdRows[1 : len(mdRows)-2]
	structuredData := make([]reasonMetadata, 0)
	for i, mdRow := range mdRows {
		if i == 0 || i == 1 {
			continue
		}
		cleanLine := strings.Split(mdRow, "|")
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

		tr := reasonMetadata{
			groupOrder:      detectGroupOrder(crState),
			crState:         crState,
			conditionType:   conditionType,
			conditionStatus: conditionStatus,
			conditionReason: conditionReason,
			remark:          remark,
		}
		structuredData = append(structuredData, tr)
	}
	return structuredData
}

func compareContent(currentTableStructured []reasonMetadata, newTableStructured []reasonMetadata) []string {
	errors := make([]string, 0)

	checkIfValuesAreSynced := func(new, old, reason string) string {
		if new != old {
			return fmt.Sprintf("Docs are not synced with Go code, difference detected in reason (%s), current value in docs is (%s) but newer in Go code is (%s)", new, old, reason)
		}
		return ""
	}

	for _, newRow := range newTableStructured {
		found := false
		for _, currentRow := range currentTableStructured {
			if newRow.conditionReason == currentRow.conditionReason {
				found = true

				if validationMessage := checkIfValuesAreSynced(currentRow.remark, newRow.remark, newRow.conditionReason); validationMessage != "" {
					errors = append(errors, validationMessage)
				}

				if validationMessage := checkIfValuesAreSynced(strconv.FormatBool(currentRow.conditionStatus), strconv.FormatBool(newRow.conditionStatus), newRow.conditionReason); validationMessage != "" {
					errors = append(errors, validationMessage)
				}

				if validationMessage := checkIfValuesAreSynced(currentRow.crState, newRow.crState, newRow.conditionReason); validationMessage != "" {
					errors = append(errors, validationMessage)
				}

				if validationMessage := checkIfValuesAreSynced(currentRow.conditionType, newRow.conditionType, newRow.conditionReason); validationMessage != "" {
					errors = append(errors, validationMessage)
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

func buildMdTable(content []reasonMetadata) string {
	renderMdElement := func(length int, content string, spaceFiller string) string {
		length = length - len(content)
		element := ""
		element += content
		for i := 0; i < length+spaceMargin; i++ {
			element += spaceFiller
		}
		return element
	}

	sort.Slice(content, func(i, j int) bool {
		if content[i].groupOrder != content[j].groupOrder {
			return content[i].groupOrder < content[j].groupOrder
		}
		return content[i].conditionReason < content[j].conditionReason
	})

	longestConditionReasons := 0
	for _, row := range content {
		l := len(row.conditionReason)
		if l > longestConditionReasons {
			longestConditionReasons = l
		}
	}

	longestRemark := 0
	for _, row := range content {
		l := len(row.remark)
		if l > longestRemark {
			longestRemark = l
		}
	}

	var mdTable string

	mdTable += fmt.Sprintf("| %s | %s | %s | %s | %s | %s | \n",
		renderMdElement(defaultMdElementSize, "No.", " "),
		renderMdElement(defaultMdElementSize, "CR state", " "),
		renderMdElement(defaultMdElementSize, "Condition type", " "),
		renderMdElement(defaultMdElementSize, "Condition status", " "),
		renderMdElement(longestConditionReasons, "Condition reason", " "),
		renderMdElement(longestRemark, "Remark", " "))

	mdTable += fmt.Sprintf("| %s | %s | %s | %s | %s | %s | \n",
		renderMdElement(defaultMdElementSize, "", "-"),
		renderMdElement(defaultMdElementSize, "", "-"),
		renderMdElement(defaultMdElementSize, "", "-"),
		renderMdElement(defaultMdElementSize, "", "-"),
		renderMdElement(longestConditionReasons, "", "-"),
		renderMdElement(longestRemark, "", "-"))

	lineNumber := 1
	for _, row := range content {
		mdTable += fmt.Sprintf("| %s | %s | %s | %s | %s | %s | \n",
			renderMdElement(defaultMdElementSize, strconv.Itoa(lineNumber), " "),
			renderMdElement(defaultMdElementSize, row.crState, " "),
			renderMdElement(defaultMdElementSize, row.conditionType, " "),
			renderMdElement(defaultMdElementSize, strconv.FormatBool(row.conditionStatus), " "),
			renderMdElement(longestConditionReasons, row.conditionReason, " "),
			renderMdElement(longestRemark, row.remark, " "))
		lineNumber++
	}

	return mdTable
}

func tryConvertGoLineToStruct(line string) (error, *reasonMetadata) {
	if line == "" {
		return fmt.Errorf("empty line given"), nil
	}
	line = strings.Replace(line, "\n", "", -1)
	parts := strings.Split(line, "//")
	if len(parts) != 2 {
		return fmt.Errorf("in line (%s) there is no comment section (//) included, comment section should have following format (//CRState;Remark)", line), nil
	}

	words := strings.Fields(parts[0])
	if len(words) != 2 {
		return fmt.Errorf("line (%s) is bad structured, it should have following format (Reason: TypeAndStatus, //CRState;Remark", line), nil
	}

	comments := strings.Split(parts[1], ";")
	if len(comments) != 2 {
		return fmt.Errorf("comment in line (%s) is bad structured, it should have following format (//CRState;Remark)", line), nil
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

	return nil, &reasonMetadata{
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
		fmt.Println(fmt.Sprintf("validation failed! -> %s", error))
	}
}

func cleanString(s *string) {
	*s = strings.Replace(*s, " ", "", -1)
	*s = strings.Replace(*s, ":", "", -1)
	*s = strings.Replace(*s, "/", "", -1)
	*s = strings.Replace(*s, ",", "", -1)
}
