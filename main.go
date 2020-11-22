package main

// Import contacts from dstar.su database to Baofeng DM-1801 contact list
// (c) 2020, EU1ADI

import (
	"flag"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"strconv"
	"strings"
)

var dstarIDs map[int]string
var newPrefixes []string

func generatePrefixes(prefixes string) {
	for _, prefix := range strings.Split(string(prefixes), ",") {
		for i := 1; i < 10; i++ {
			newPrefixes = append(newPrefixes, fmt.Sprintf("%s%d", prefix, i))
		}
	}
}

func checkPrefix(ID string) bool {
	for _, prefix := range newPrefixes {
		if strings.HasPrefix(ID, prefix) {
			return true
		}
	}
	return false
}

func getIDs() {
	resp, err := http.Get("http://registry.dstar.su/dmr/DMRIds.php")
	if err != nil {
		return
	}

	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return
	}

	dstarIDs = make(map[int]string)
	ids := strings.Split(string(body), "\n")

	for _, line := range ids {
		if len(line) > 0 && line[:1] == "#" { // skip commented line
			continue
		}

		record := strings.Split(line, "\t")

		if checkPrefix(record[0]) && len(record) > 1 {
			alias := record[2]
			id, _ := strconv.Atoi(strings.TrimSpace(record[0]))
			dstarIDs[id] = alias
		}
	}
}

func checkID(contacts []string, ID int) bool {
	for _, line := range contacts {
		if len(line) > 0 && (line[:1] == "#" || line[:1] == "N") { // skip header or commented line
			continue
		}

		contact := strings.Split(line, ",")

		if len(contact) > 3 {
			contactID, err := strconv.Atoi(strings.TrimSpace(contact[2]))
			if err != nil {
				continue
			}

			if contactID == ID {
				return true
			}
		}
	}

	return false
}

func processContacts(fileName string) {
	content, err := ioutil.ReadFile(fileName)
	if err != nil {
		return
	}

	contacts := strings.Split(string(content), "\n")
	newIDs := make(map[int]string)

	lastLine := strings.Split(contacts[len(contacts)-2], ",")
	lineNumber, _ := strconv.Atoi(strings.TrimSpace(lastLine[0]))

	for dstarID, dstarName := range dstarIDs {
		if !checkID(contacts, dstarID) {
			newIDs[dstarID] = dstarName
		}
	}

	newLines := []string{}
	for dstarID, dstarName := range newIDs {
		lineNumber = lineNumber + 1
		newLines = append(newLines, fmt.Sprintf("%d,%s,%d,Private Call,Off,None,None", lineNumber, dstarName, dstarID))
	}

	f, err := os.OpenFile(fileName, os.O_APPEND|os.O_WRONLY, 0644)
	if err != nil {
		return
	}

	for _, line := range newLines {
		println(line)
		f.WriteString(line + "\n")
	}
	defer f.Close()
}

func main() {
	prefix := flag.String("prefix", "257", "DMRID country prefix")
	flag.Parse()
	tail := flag.Args()

	fileName := tail[0]

	if fileName != "" {
		generatePrefixes(*prefix)
		getIDs()
		processContacts(fileName)
	}
}
