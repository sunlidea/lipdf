package core

import (
	"encoding/json"
	"fmt"
	"github.com/google/uuid"
	"io/ioutil"
	"os"
	"testing"
)

// test generate fdf
func TestGenerateFdfByPdf(t *testing.T) {
	pdfPath := "../file/1022.pdf"
	fdfPath := "../file/1022.fdf"

	err := GenerateFdf(pdfPath, fdfPath)
	if err != nil {
		t.Fatalf("fail to GenerateFdfByPdf:%v", err)
		return
	}
}

// test readFormFields
func TestReadFormFields(t *testing.T) {
	fdfPath := "../file/1022.fdf"
	keys, err := readFormFields(fdfPath)
	if err != nil {
		t.Fatalf("fail to readFormFields:%v", err)
		return
	}
	t.Logf("%+v\n", keys)
}

func TestPdfFormFields(t *testing.T) {
	pdfPath := "../file/1022.pdf"
	resultData, err := pdfFormFields(pdfPath)
	if err != nil {
		t.Fatalf("fail to pdfFormFields:%v", err)
		return
	}
	t.Logf("%+v\n", resultData)

	result := make(map[string]interface{})
	for k, _ := range resultData {
		result[k] = k
	}
	data, err := json.Marshal(result)
	if err != nil {
		t.Fatalf("fail to pdfFormFields:%v", err)
		return
	}
	f, err := os.Create("../file/result.json")
	if err != nil {
		t.Fatalf("Create:%v", err)
		return
	}
	defer f.Close()

	_, err = f.Write(data)
	if err != nil {
		t.Fatalf("Write:%v", err)
		return
	}
}

func TestFillForm(t *testing.T) {
	pdfPath := "../file/1022.pdf"
	data, err := ioutil.ReadFile("../file/result.json")
	if err != nil {
		t.Fatalf("Open:%v", err)
		return
	}

	var m map[string]interface{}
	err = json.Unmarshal(data, &m)
	if err != nil {
		t.Fatalf("Unmarshal:%v", err)
		return
	}

	_, err = FillForm(m, pdfPath, true)
	if err != nil {
		t.Fatalf("FillForm:%v", err)
		return
	}
}

func TestDumpFields(t *testing.T) {
	pdfPath := "../file/1022.pdf"
	fileID := uuid.New()
	// dump fields to dest file
	dumpPath := fmt.Sprintf("../file/%s.dump", fileID)
	err := dumpFields(pdfPath, dumpPath)
	if err != nil {
		t.Fatalf("dumpFields:%v", err)
		return
	}
	//defer os.Remove(dumpPath)

	// read dump fields
	fields, err := readDumpFields(dumpPath)
	if err != nil {
		t.Fatalf("readDumpFields:%v", err)
		return
	}

	t.Logf("%+v\n", fields)
}

func TestPdfFieldsToJSON(t *testing.T) {
	_, err := PdfFieldsToJSON("../file/1022.pdf")
	if err != nil {
		t.Fatalf("PdfFieldsToJSON:%v", err)
		return
	}
}
