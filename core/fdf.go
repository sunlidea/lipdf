package core

import (
	"bufio"
	"container/list"
	"context"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"
)

func init() {
	// Check if the pdftk utility Exists.
	_, err := exec.LookPath("pdftk")
	if err != nil {
		panic("pdftk utility is not installed!")
	}
}

//generate fdf file from pdf
func GenerateFdf(pdfPath string, destPath string) (err error) {
	err = generateCore(pdfPath, destPath, []string{"generate_fdf"})
	if err != nil {
		return fmt.Errorf("failed to invoke generateCore: %v", err)
	}
	return nil
}

// exec pdftk
func generateCore(pdfPath string, destPath string, options []string) (err error) {
	pdfPath, err = filepath.Abs(pdfPath)
	if err != nil {
		return fmt.Errorf("filepath abs fail|%v|%s", err, pdfPath)
	}

	destPath, err = filepath.Abs(destPath)
	if err != nil {
		return fmt.Errorf("filepath abs fail|%v|%s", err, destPath)
	}

	// Check if the form file Exists.
	e, err := Exists(pdfPath)
	if err != nil {
		return fmt.Errorf("check pdf file Exists fail: %v", err)
	} else if !e {
		return fmt.Errorf("pdf file does not Exists: '%s'", pdfPath)
	}

	// Create a temporary directory.
	tmpDir, err := ioutil.TempDir("", "dest-")
	if err != nil {
		return fmt.Errorf("create temporary directory fail: %v", err)
	}

	// Remove the temporary directory on defer again.
	defer func() {
		errD := os.RemoveAll(tmpDir)
		// Log the error only.
		if errD != nil {
			log.Printf("fillpdf: failed to remove temporary directory '%s' again: %v", tmpDir, errD)
		}
	}()

	// Create the temporary output file path.
	outFdfFile := filepath.Clean(tmpDir + "/output")

	// Check if the destination file Exists.
	e, err = Exists(destPath)
	if err != nil {
		return fmt.Errorf("failed to check if destination PDF file Exists: %v", err)
	} else if e {
		err = os.Remove(destPath)
		if err != nil {
			return fmt.Errorf("failed to remove destination PDF file: %v", err)
		}
	}

	//generate fdf file command args
	args := make([]string, 0, 5)
	//input file
	args = append(args, pdfPath)
	//options
	args = append(args, options...)
	//output file
	args = append(args, "output", outFdfFile)

	ctx, _ := context.WithTimeout(context.Background(), time.Second*120)
	_, err = execCmdInDir(ctx, tmpDir, "pdftk", args...)
	if err != nil {
		return fmt.Errorf("pdftk exec fail: %v", err)
	}

	// On success, copy the output file to the final destination.
	err = copyFile(outFdfFile, destPath)
	if err != nil {
		return fmt.Errorf("failed to copy created output file to final destination: %v", err)
	}

	return nil
}

// read and parse pdf form field keys
func readFormFields(filePath string) (map[string]struct{}, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return nil, fmt.Errorf("fail to open file:%v", err)
	}
	defer file.Close()

	reader := bufio.NewReader(file)

	l := list.New()
	var line string
	for {
		line, err = reader.ReadString('\n')
		if err != nil {
			break
		}

		if strings.Contains(line, "[")  {
			l.PushBack("[")
		}
		if strings.Contains(line, "]")  {
			l.PushBack("]")
		}

		if strings.HasPrefix(line, "/T")  {
			l.PushBack(strings.TrimSuffix(strings.TrimPrefix(line, "/T ("), ")\n"))
		}
	}
	if err != io.EOF {
		return nil, fmt.Errorf("fail to read file:%v", err)
	}

	if l.Len() < 2 {
		return nil, nil
	}

	//trim first "["
	l.Remove(l.Front())
	//trim last "]"
	l.Remove(l.Back())

	keys := make(map[string]struct{})
	prefixes := make([]string, 0, 1)
	for l.Len() > 0 {
		str := l.Back().Value.(string)
		if str != "]" && str != "[" {

			//prev value
			prev := l.Back().Prev()
			if prev == nil {
				keys[str] = struct{}{}
				break
			}

			//last prefix
			prefix := ""
			if len(prefixes) > 0 {
				prefix = prefixes[len(prefixes)-1]
			}

			prevStr:= prev.Value.(string)
			if prevStr == "]" {
				//just prefix, don't need to add to keys
				if len(prefix) > 0 {
					prefix = fmt.Sprintf("%s.%s", prefix, str)
				}else {
					prefix = str
				}
				prefixes = append(prefixes, prefix)
			}else {

				// add to keys
				k := ""
				if len(prefix) > 0 {
					k = fmt.Sprintf("%s.%s", prefix, str)
				}else {
					k = str
				}
				keys[k] = struct{}{}

				// [[[a]b]c]
				if prevStr == "[" {
					for prevStr == "[" {
						prefixes = prefixes[0:len(prefixes)-1]
						prev = prev.Prev()
						if prev != nil {
							prevStr = prev.Value.(string)
						}else {
							prevStr = ""
						}
					}
				}
			}
		}
		l.Remove(l.Back())
	}

	return keys, nil
}



