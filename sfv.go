// Package sfv provides a simple way of reading and verifying SFV (Simple File
// Verification) files.
package sfv

import (
	"bufio"
	"fmt"
	"hash/crc32"
	"io"
	"os"
	"path"
	"strconv"
	"strings"
)

// Checksum represents a line in a SFV file, containing the filename, full path
// to the file and the CRC32 checksum
type Checksum struct {
	Filename string
	Path     string
	CRC32    uint32
}

// SFV contains all the checksums read from a SFV file.
type SFV struct {
	Checksums []Checksum
}

// Verify calculates the CRC32 of the associated file and returns true if the
// checksum is correct.
func (c *Checksum) Verify() (bool, error) {
	f, err := os.Open(c.Path)
	if err != nil {
		return false, err
	}
	defer f.Close()

	h := crc32.NewIEEE()
	reader := bufio.NewReader(f)
	buf := make([]byte, 4096)
	for {
		n, err := reader.Read(buf)
		if err != nil && err != io.EOF {
			return false, err
		}
		if n == 0 {
			break
		}
		h.Write(buf[:n])
	}
	success := h.Sum32() == c.CRC32
	return success, nil
}

// Verify verifies all checksums contained in SFV and returns true if all
// checksums are correct.
func (c *SFV) Verify() (bool, error) {
	for _, c := range c.Checksums {
		ok, err := c.Verify()
		if err != nil {
			return false, err
		}
		if !ok {
			return false, nil
		}
	}
	return true, nil
}

func parseChecksum(dir string, s string) (*Checksum, error) {
	words := strings.SplitN(s, " ", 2)
	if len(words) != 2 {
		return nil, fmt.Errorf("expected 2 words, got %d", len(words))
	}

	filename := words[0]
	path := path.Join(dir, filename)
	crcString := strings.TrimLeft(words[1], " \t")
	crc32, err := strconv.ParseUint(crcString, 16, 32)
	if err != nil {
		return nil, err
	}
	// ParseUint will return error if number exceeds 32 bits
	return &Checksum{
		Path:     path,
		Filename: filename,
		CRC32:    uint32(crc32),
	}, nil
}

// Read reades a SFV file from filepath and creates a new SFV containing
// checksums parsed from the SFV file.
func Read(filepath string) (*SFV, error) {
	f, err := os.Open(filepath)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	dir := path.Dir(filepath)
	checksums := []Checksum{}
	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		line := strings.Trim(scanner.Text(), " \t")
		if len(line) == 0 || strings.HasPrefix(line, ";") {
			continue
		}
		checksum, err := parseChecksum(dir, line)
		if err != nil {
			return nil, err
		}
		checksums = append(checksums, *checksum)
	}

	if err := scanner.Err(); err != nil {
		return nil, err
	}

	return &SFV{Checksums: checksums}, nil
}
