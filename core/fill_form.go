package core

import (
	"bufio"
	"fmt"
	"github.com/google/uuid"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"strings"
)

type Field struct {
	FieldType string `json:"FieldType"`
	FieldName string `json:"FieldName"`
	ViewName  string `json:"ViewName"`
	FieldOptions []string `json:"FieldOptions,omitempty"`
}

type GroupField struct {
	Fields []Field       `json:"Fields"`
	GroupName string     `json:"GroupName"`
	ShowName string      `json:"ShowName"`
}

type FieldInfo struct {
	PdfPath      string       `json:"PdfPath"`
	GroupFields  []GroupField `json:"GroupFields"`
	SingleFields []Field      `json:"SingleFields"`
}

// extract form fields and convert to json
func PdfFieldsToJSON(pdfPath string) (*FieldInfo, error) {
	rawFields, err := pdfFormFields(pdfPath)
	if err != nil {
		return nil, fmt.Errorf("fail to pdfFormFields: %+v", err)
	}

	tmpFields := make(map[string]*GroupField, len(rawFields))
	for k, fd := range rawFields {

		ks := strings.Split(k, " ")
		gKey := ks[0]
		if _, ok := tmpFields[gKey]; !ok {
			g := &GroupField{
				Fields: make([]Field, 0, 1),
				GroupName: gKey,
			}
			tmpFields[gKey] = g
		}

		g := tmpFields[gKey]
		g.Fields = append(g.Fields, fd)
		tmpFields[gKey] = g
	}

	groupFields := make([]GroupField, 0, len(tmpFields))
	singleFields := make([]Field, 0, len(tmpFields))
	for _, gf := range tmpFields {
		if len(gf.Fields) > 1 {
			groupFields = append(groupFields, *gf)
		}else if len(gf.Fields) == 1 {
			singleFields = append(singleFields, gf.Fields[0])
		}
	}
	result := &FieldInfo{
		PdfPath: pdfPath,
		GroupFields:groupFields,
		SingleFields:singleFields,
	}

	return result, nil
}

// extract pdf form infos
func pdfFormFields(pdfPath string) (map[string]Field, error) {
	fileID := uuid.New()

	// dump fields to dest file
	dumpPath, err := filepath.Abs(fmt.Sprintf("file/%s.dump", fileID))
	if err != nil {
		return nil, err
	}
	err = dumpFields(pdfPath, dumpPath)
	if err != nil {
		return nil, err
	}
	defer os.Remove(dumpPath)

	// read dump fields
	fields, err := readDumpFields(dumpPath)
	if err != nil {
		return nil, err
	}

	// generate fdf file
	fdfPath, err := filepath.Abs(fmt.Sprintf("file/%s.fdf", fileID))
	if err != nil {
		return nil, err
	}
	err = GenerateFdf(pdfPath, fdfPath)
	if err != nil {
		return nil, err
	}
	defer os.Remove(fdfPath)

	// pdf form keys
	formKeys, err := readFormFields(fdfPath)
	if err != nil {
		return nil, err
	}

	// select form fields from all fields
	result := make(map[string]Field)
	for k, v := range fields {
		if _, ok := formKeys[k]; ok {
			result[k] = v
		}
	}

	return result, nil
}

// fill form to designated pdf
func FillForm(form map[string]interface{}, pdfPath string) (string, error) {

	// Create a temporary directory.
	tmpDir, err := ioutil.TempDir("", "fillpdf-")
	if err != nil {
		return "", fmt.Errorf("failed to create temporary directory: %v", err)
	}

	// Remove the temporary directory on defer again.
	defer func() {
		errD := os.RemoveAll(tmpDir)
		// Log the error only.
		if errD != nil {
			log.Printf("fillpdf: failed to remove temporary directory '%s' again: %v", tmpDir, errD)
		}
	}()

	// Create the fdf data file.
	fdfFile := filepath.Clean(tmpDir + "/data.fdf")
	err = createFdfFile(form, fdfFile)
	if err != nil {
		return "", fmt.Errorf("failed to create fdf form data file: %v", err)
	}

	outID := fmt.Sprintf("%s.pdf", uuid.New())
	outPdfPath := fmt.Sprintf("../file/%s", outID)

	// pdftk form.pdf fill_form data.fdf output form.filled.pdf
	args := []string {
		"fill_form",
		fdfFile,
	}
	err = generateCore(pdfPath, outPdfPath, args)
	if err != nil {
		return "", fmt.Errorf("pdftk exec fail: %v", err)
	}
	return outPdfPath, nil
}

func createFdfFile(form map[string]interface{}, path string) error {
	// Create the file.
	file, err := os.Create(path)
	if err != nil {
		return err
	}
	defer file.Close()

	// Create a new writer.
	w := bufio.NewWriter(file)

	// Write the fdf header.
	fmt.Fprintln(w, fdfHeader)

	// Write the form data.
	for key, value := range form {
		fmt.Fprintf(w, "<< /T (%s) /V (%v)>>\n", key, value)
	}

	// Write the fdf footer.
	fmt.Fprintln(w, fdfFooter)

	// Flush everything.
	return w.Flush()
}

const fdfHeader = `%FDF-1.2
%,,oe"
1 0 obj
<<
/FDF << /Fields [`

const fdfFooter = `]
>>
>>
endobj
trailer
<<
/Root 1 0 R
>>
%%EOF`

