package utils

import (
	"bufio"
	"context"
	"io"
	"math/rand"
	"os"
	"os/exec"
	"time"
)

func CreateDirectories(directories ...string) error {
	for _, path := range directories {
		_, err := os.Stat(path)
		if os.IsNotExist(err) {
			err := os.MkdirAll(path, 0755)
			if err != nil {
				return err
			}
		} else if err != nil {
			return err
		}
	}

	return nil
}

func CreateFile(path, contents string) error {
	f, err := os.Create(path)
	if err != nil {
		return err
	}

	_, err = f.WriteString(contents)
	if err != nil {
		return err
	}

	return nil
}

// GenerateRandString generates a variable length string; can be used for ids
func GenerateRandString(length int) string {

	const charset = "abcdefghijklmnopqrstuvwxyz" +
		"ABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"

	var seededRand = rand.New(
		rand.NewSource(time.Now().UnixNano()))

	b := make([]byte, length)
	for i := range b {
		b[i] = charset[seededRand.Intn(len(charset))]
	}

	return string(b)
}

// ExecuteCmd wraps context around a given command and executes it.
func ExecuteCmd(path string, args []string, env []string, dir string) ([]byte, error) {
	// Create context with timeout
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Minute)
	defer cancel()

	// Create command
	cmd := exec.CommandContext(ctx, path, args...)
	cmd.Env = env
	cmd.Dir = dir

	// Execute command
	return cmd.CombinedOutput()
}

func ReadLine(r io.Reader, lineNum int) (line string, lastLine int, err error) {
	sc := bufio.NewScanner(r)
	for sc.Scan() {
		lastLine++
		if lastLine == lineNum {
			// you can return sc.Bytes() if you need output in []bytes
			return sc.Text(), lastLine, sc.Err()
		}
	}
	return line, lastLine, io.EOF
}
