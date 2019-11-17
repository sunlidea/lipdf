package handler

import (
	"encoding/json"
	"fmt"
	"github.com/google/uuid"
	"github.com/labstack/echo"
	"github.com/sunlidea/lipdf/core"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"path/filepath"
	"strings"
)

type WebHandler struct {}

// example
func (wh *WebHandler) Example(c echo.Context) error {

	pdfPath := c.FormValue("PdfPath")
	absPdfPath, err := filepath.Abs(pdfPath)
	if err != nil {
		return c.NoContent(http.StatusBadRequest)
	}
	e, err := core.Exists(absPdfPath)
	if err != nil || !e{
		return c.NoContent(http.StatusBadRequest)
	}

	e, err = core.Exists(absPdfPath)
	if err != nil {
		return c.NoContent(http.StatusBadRequest)
	} else if !e {
		return c.NoContent(http.StatusBadRequest)
	}

	jsonPath := fmt.Sprintf("%s.json", strings.TrimSuffix(absPdfPath, ".pdf"))
	e, err = core.Exists(jsonPath)
	if err != nil || !e{
		return c.NoContent(http.StatusBadRequest)
	}

	data, err := ioutil.ReadFile(jsonPath)
	if err != nil {
		return c.NoContent(http.StatusInternalServerError)
	}
	var fieldInfo core.FieldInfo
	err = json.Unmarshal(data, &fieldInfo)
	if err != nil {
		return c.NoContent(http.StatusInternalServerError)
	}
	fieldInfo.PdfPath = pdfPath

	return c.JSON(http.StatusOK, fieldInfo)
}

// submit fileds, fill form
func (wh *WebHandler) Submit(c echo.Context) error {

	var m map[string]interface{}
	params := c.FormValue("Fields")
	err := json.Unmarshal([]byte(params), &m)
	if err != nil {
		return c.NoContent(http.StatusBadRequest)
	}

	pdfPath, err := filepath.Abs(c.FormValue("PdfPath"))
	if err != nil {
		return c.NoContent(http.StatusBadRequest)
	}
	e, err := core.Exists(pdfPath)
	if err != nil || !e{
		return c.NoContent(http.StatusBadRequest)
	}

	outPath, err := core.FillForm(m, pdfPath)
	if err != nil {
		return c.NoContent(http.StatusInternalServerError)
	}

	return c.String(http.StatusOK, outPath)
}

//upload file
func (wh *WebHandler) Upload(c echo.Context) error {

	// Source
	file, err := c.FormFile("file")
	if err != nil {
		return c.NoContent(http.StatusBadRequest)
	}
	src, err := file.Open()
	if err != nil {
		return c.NoContent(http.StatusBadRequest)
	}
	defer src.Close()

	// Destination
	fileID := uuid.New()
	path := fmt.Sprintf("file/%s.pdf", fileID)
	absPath, err := filepath.Abs(path)
	if err != nil {
		return c.NoContent(http.StatusBadRequest)
	}

	dst, err := os.Create(absPath)
	if err != nil {
		return c.NoContent(http.StatusInternalServerError)
	}
	defer dst.Close()

	// Copy
	if _, err = io.Copy(dst, src); err != nil {
		return c.NoContent(http.StatusInternalServerError)
	}

	//convert to json
	fileInfo, err := core.PdfFieldsToJSON(path)
	if err != nil {
		return c.NoContent(http.StatusInternalServerError)
	}

	return c.JSON(http.StatusOK, fileInfo)
}
