package core

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"strings"
)

//dump field data from pdf
func dumpFields(pdfPath string, destPath string) (err error) {
	err = generateCore(pdfPath, destPath, []string{"dump_data_fields_utf8"})
	if err != nil {
		return fmt.Errorf("failed to invoke generateCore: %v", err)
	}
	return nil
}

// read and parse dump field data
func readDumpFields(filePath string) (map[string]Field, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return nil, fmt.Errorf("fail to open file:%v", err)
	}
	defer file.Close()

	reader := bufio.NewReader(file)

	fields := make(map[string]Field)
	fd := Field{}
	var line string
	for {
		line, err = reader.ReadString('\n')
		if err != nil {
			break
		}

		// new field info
		if strings.HasSuffix(line, "---\n") {
			if len(fd.FieldName) > 0{
				fields[fd.FieldName] = fd
			}
			fd = Field{}
		}

		if strings.HasPrefix(line, "FieldType") {
			strs := strings.Split(strings.TrimSuffix(line, "\n"), ": ")
			if len(strs) > 1 {
				fd.FieldType = strs[1]
			}
		}

		if strings.HasPrefix(line, "FieldName") {
			strs := strings.Split(strings.TrimSuffix(line, "\n"), ": ")
			if len(strs) > 1 {
				fd.FieldName = strs[1]
			}
		}

		if strings.HasPrefix(line, "FieldStateOption") {
			strs := strings.Split(strings.TrimSuffix(line, "\n"), ": ")
			if len(strs) > 1 {
				fd.FieldOptions = append(fd.FieldOptions, strs[1])
			}
		}
	}
	if err != io.EOF {
		return nil, fmt.Errorf("fail to read file:%v", err)
	}
	// last field
	if len(fd.FieldName) > 0{
		fields[fd.FieldName] = fd
	}

	return fields, nil
}
