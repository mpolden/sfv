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
	Path      string
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
	return h.Sum32() == c.CRC32, nil
}

// IsExist returns a boolean indicating if the file associated with the checksum
// exists
func (c *Checksum) IsExist() bool {
	_, err := os.Stat(c.Path)
	return err == nil
}

// Verify verifies all checksums contained in SFV and returns true if all
// checksums are correct.
func (s *SFV) Verify() (bool, error) {
	if len(s.Checksums) == 0 {
		return false, fmt.Errorf("no checksums found in %s", s.Path)
	}
	for _, c := range s.Checksums {
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

// IsExist returns a boolean if all the files in SFV exists
func (s *SFV) IsExist() bool {
	for _, c := range s.Checksums {
		if !c.IsExist() {
			return false
		}
	}
	return true
}

func parseChecksum(dir string, line string) (*Checksum, error) {
	parts := strings.SplitN(line, " ", 2)
	if len(parts) != 2 {
		return nil, fmt.Errorf("could not parse checksum: %q", line)
	}
	filename := strings.TrimSpace(parts[0])
	path := path.Join(dir, filename)
	crc32, err := strconv.ParseUint(strings.TrimSpace(parts[1]), 16, 32)
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

func parseChecksums(dir string, r io.Reader) ([]Checksum, error) {
	checksums := []Checksum{}
	scanner := bufio.NewScanner(r)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
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
	return checksums, nil
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
	checksums, err := parseChecksums(dir, f)
	if err != nil {
		return nil, err
	}
	return &SFV{
		Checksums: checksums,
		Path:      filepath,
	}, nil
}
