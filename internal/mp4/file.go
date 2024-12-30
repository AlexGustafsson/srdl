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
func (f *File) SeekAtom(needle string) (int64, uint32, error) {
	offset, err := f.Seek(0, io.SeekCurrent)
	if err != nil {
		return -1, 0, err
	}

	var relativeOffset int64 = 0
	for {
		atomSize, atomType, err := f.ReadAtomHeader()
		if err != nil {
			return -1, 0, err
		}

		if atomType == needle {
			return offset + relativeOffset, atomSize, nil
		}

		// Seek past the atom's data
		if _, err := f.Seek(int64(atomSize)-8, io.SeekCurrent); err != nil {
			return -1, 0, err
		}

		relativeOffset += int64(atomSize)
	}
}

// Allocate allocates a chunk of size bytes at offset.
// That is, it truncates the underlying file and "moves" (copies) the data at
// offset size bytes further in.
func (f *File) Allocate(offset int64, size int64) error {
	fileOffset, err := f.Seek(0, io.SeekCurrent)
	if err != nil {
		return err
	}

	// Keep the current offset even if we allocate a gap before it
	if fileOffset > offset {
		fileOffset += size
	}

	stat, err := f.Stat()
	if err != nil {
		return err
	}

	if err := f.Truncate(stat.Size() + size); err != nil {
		return err
	}

	_, err = f.Seek(offset+size, io.SeekStart)
	if err != nil {
		return err
	}

	readFile, err := os.Open(f.Name())
	if err != nil {
		return err
	}

	_, err = readFile.Seek(offset, io.SeekStart)
	if err != nil {
		readFile.Close()
		return err
	}

	reader := io.LimitReader(readFile, stat.Size()-offset)

	_, err = io.Copy(f, reader)
	if err != nil {
		readFile.Close()
		return err
	}

	_, err = readFile.Seek(fileOffset, io.SeekStart)
	if err != nil {
		readFile.Close()
		return err
	}

	return readFile.Close()
}

// ReadAtomHeader reads the header of an atom at the current location.
// Returns its size and type.
func (f *File) ReadAtomHeader() (uint32, string, error) {
	var buffer [8]byte
	if _, err := f.Read(buffer[:]); err != nil {
		return 0, "", err
	}

	return binary.BigEndian.Uint32(buffer[0:4]), string(buffer[4:8]), nil
}

func (f *File) ReadAtomHeaderAt(offset int64) (uint32, string, error) {
	var buffer [8]byte
	if _, err := f.ReadAt(buffer[:], offset); err != nil {
		return 0, "", err
	}

	return binary.BigEndian.Uint32(buffer[0:4]), string(buffer[4:8]), nil
}

func (f *File) WriteAtomHeader(atomSize uint32, atomType string) error {
	return writeAtomHeader(f, atomSize, atomType)
}

func (f *File) WriteAtom(atomSize uint32, atomType string, data []byte) error {
	if err := f.WriteAtomHeader(atomSize, atomType); err != nil {
		return err
	}

	_, err := f.Write(data)
	return err
}

func (f *File) WriteAtomHeaderAt(atomSize uint32, atomType string, offset int64) error {
	buffer := formatAtomHeader(atomSize, atomType)
	_, err := f.WriteAt(buffer[:], offset)
	return err
}

func (f *File) ReadAtomAt(offset int64) (uint32, string, []byte, error) {
	size, atom, err := f.ReadAtomHeaderAt(offset)
	if err != nil {
		return 0, "", nil, err
	}

	data := make([]byte, size)
	if _, err := f.ReadAt(data, offset); err != nil {
		return 0, "", nil, err
	}

	return size, atom, data, nil
}

func (f *File) WriteAtomAt(atomSize uint32, atomType string, data []byte, offset int64) error {
	if err := f.WriteAtomHeaderAt(atomSize, atomType, offset); err != nil {
		return err
	}

	_, err := f.WriteAt(data, offset+8)
	return err
}

func (f *File) CopyAtom(from int64, to int64) error {
	size, atom, data, err := f.ReadAtomAt(from)
	if err != nil {
		return err
	}

	return f.WriteAtomAt(size, atom, data, to)
}

func formatAtomHeader(atomSize uint32, atomType string) []byte {
	var buffer [8]byte
	binary.BigEndian.PutUint32(buffer[0:], atomSize)
	copy(buffer[4:], atomType)
	return buffer[:]
}

func writeAtomHeader(w io.Writer, atomSize uint32, atomType string) error {
	buffer := formatAtomHeader(atomSize, atomType)
	_, err := w.Write(buffer[:])
	return err
}
