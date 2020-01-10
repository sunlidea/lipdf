package core

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"os"
	"os/exec"
)

// Run the command in the specified Dir
func execCmdInDir(ctx context.Context, dir, name string, args ...string) ([]byte, error) {
	//output after cmd exec
	var outputBuf bytes.Buffer
	cmd := exec.Command(name, args...)
	cmd.Stdout = &outputBuf
	cmd.Dir = dir

	//start
	err := cmd.Start()
	if err != nil {
		return nil, err
	}

	//wait for cmd exec
	done := make(chan error, 1)
	go func() {
		done <- cmd.Wait()
	}()

	select {
	case <-ctx.Done():
		err = cmd.Process.Kill()
		if err != nil {
			return nil, err
		}
		return nil, fmt.Errorf("exec time out")
	case <-done:
		//cmd exec finish
	}

	return outputBuf.Bytes(), nil
}

// Exists returns whether the given file or directory Exists or not
func Exists(path string) (bool, error) {
	_, err := os.Stat(path)
	if err == nil {
		return true, nil
	}
	if os.IsNotExist(err) {
		return false, nil
	}
	return false, err
}

// copyFile copies the contents of the file named src to the file named
// by dst. The file will be created if it does not already exist. If the
// destination file Exists, all it's contents will be replaced by the contents
// of the source file.
func copyFile(src, dst string) (err error) {
	in, err := os.Open(src)
	if err != nil {
		return
	}
	defer in.Close()
	out, err := os.Create(dst)
	if err != nil {
		return
	}
	defer func() {
		cerr := out.Close()
		if err == nil {
			err = cerr
		}
	}()
	if _, err = io.Copy(out, in); err != nil {
		return
	}
	err = out.Sync()
	return
}
