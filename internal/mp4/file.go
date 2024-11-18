package mp4

import (
	"encoding/binary"
	"io"
	"os"
)

const HeaderSize = 8

// File provides MP4 operations on a file.
//
// NOTE: It's not meant to be the most performant solution possible to modifying
// MP4s, but it's designed to strike a nice middle ground. It should still be
// very performant, under the assumption that underlying file operations are
// cheap and that the file isn't huge.
type File struct {
	*os.File
}

// SeekAtom seeks for atom in the current level of the MP4 hierarchy.
// Returns the offset and size of the first atom that matches.
// The file will be positioned after the atom's header.
func (f *File) SeekAtom(atom string) (int64, uint32, error) {
	offset, err := f.Seek(0, io.SeekCurrent)
	if err != nil {
		return -1, 0, err
	}

	var boxOffset uint32 = 0
	for {
		boxSize, boxType, err := f.ReadBoxHeader()
		if err == io.EOF {
			return -1, 0, nil
		} else if err != nil {
			return -1, 0, err
		}

		if boxType == atom {
			return offset + int64(boxOffset), boxSize, nil
		}

		// Seek past the box's data
		if _, err := f.Seek(int64(boxSize)-8, io.SeekCurrent); err != nil {
			return -1, 0, err
		}

		boxOffset += boxSize
	}
}

func (f *File) ReadBoxHeader() (uint32, string, error) {
	var buffer [8]byte
	if _, err := f.Read(buffer[:]); err != nil {
		return 0, "", err
	}

	return binary.BigEndian.Uint32(buffer[0:4]), string(buffer[4:8]), nil
}

func (f *File) WriteBoxHeader(boxSize uint32, boxType string) error {
	return writeBoxHeader(f, boxSize, boxType)
}

func formatBoxHeader(boxSize uint32, boxType string) []byte {
	var buffer [8]byte
	binary.BigEndian.PutUint32(buffer[0:], boxSize)
	copy(buffer[4:], []byte(boxType))
	return buffer[:]
}

func writeBoxHeader(w io.Writer, boxSize uint32, boxType string) error {
	buffer := formatBoxHeader(boxSize, boxType)
	_, err := w.Write(buffer[:])
	return err
}
