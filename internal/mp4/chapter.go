package mp4

import (
	"bytes"
	"encoding/binary"
	"io"
	"os"
)

type Index struct {
	Chapters []Chapter
}

func (i Index) Bytes() []byte {
	var buffer bytes.Buffer
	for _, chapter := range i.Chapters {
		buffer.Write(chapter.Bytes())
	}
	return buffer.Bytes()
}

type Chapter struct {
	Name string
}

func (c Chapter) Bytes() []byte {
	// TODO: Is the first byte just a tail | head error?
	// Each "frame" is a chapter. Its contents is the name of a chapter as a
	// pascal string, preceeded by two bytes for its size.
	// The frame is preceeded by an unknown byte and is folloed by an encoding
	// "atom":
	// 00 00 00 0c 65 6e 63 64 00 00 01
	//  ?  ?  ?  ?  e  n  c  d 00 00 01 (UTF-8)

	buffer := make([]byte, 2+len(c.Name)+12)
	binary.BigEndian.PutUint16(buffer[0:2], uint16(len(c.Name)))
	copy(buffer[2:], c.Name)
	copy(buffer[2+len(c.Name):], "\x00\x00\x00\x0c\x65encd\x00\x00\x01")

	return buffer
}

// Write ...
//
// SEE: https://developer.apple.com/standards/qtff-2001.pdf pp. 143
//
// SEE: http://forum.doom9.org/archive/index.php/t-166802.html
//
// SEE: https://mp4ra.org/registered-types/boxes
//
// SEE: https://yhf8377.medium.com/quicktime-movie-editing-using-avfoundation-and-swift-5-33965c522abc
func (i Index) Write(f *os.File) error {
	mp4 := File{f}

	moovOffset, moovSize, err := mp4.SeekAtom("moov")
	if err != nil {
		return err
	}

	trakOffset, trakSize, err := mp4.SeekAtom("trak")
	if err != nil {
		return err
	}

	// Assume no tref atom exists, just create it at the end of the trak box
	trefOffset := trakOffset + int64(trakSize)
	if err := mp4.Allocate(trefOffset, 20); err != nil {
		return err
	}

	// Write the tref atom
	_, err = mp4.Seek(trefOffset, io.SeekStart)
	if err != nil {
		return err
	}

	if err := mp4.WriteAtomHeader(20, "tref"); err != nil {
		return err
	}

	// Write the chap atom
	if err := mp4.WriteAtomHeader(12, "chap"); err != nil {
		return err
	}
	// Write a reference to the second trak
	if _, err := mp4.Write([]byte{0x00, 0x00, 0x00, 0x02}); err != nil {
		return err
	}

	moovSize += 20
	trakSize += 20

	// Update parent sizes
	if err := mp4.WriteAtomHeaderAt(moovSize, "moov", moovOffset); err != nil {
		return err
	}
	if err := mp4.WriteAtomHeaderAt(trakSize, "trak", trakOffset); err != nil {
		return err
	}

	// TODO: This part of the tree is quite complex - write in reverse instead
	// in order to more easily have size for headers?
	// // Allocate room for a second trak after the first
	// // TODO: Actual size
	// // TODO: Update parent (moov) size
	// trak2Offset := trakOffset + int64(trakSize) + 20
	// if err := mp4.Allocate(trak2Offset, 100); err != nil {
	// 	return err
	// }

	// // Create a second trak
	// if err := mp4.WriteAtomHeaderAt(100, "trak", trak2Offset); err != nil {
	// 	return err
	// }

	// // Copy the tkhd from the first trak, assuming it's the first atom
	// if err := mp4.CopyAtom(trakOffset+8, trak2Offset+8); err != nil {
	// 	return err
	// }

	// Write an mdat atom for housing our chapters
	// Seems to be flag thingy with single byte for size - two bytes (flag??) - size - text

	// NOTE: It seems as if SR's MP4 files have their actual data after all atoms,
	// not in an mdat block.

	mdat := i.Bytes()
	if err := mp4.Allocate(moovOffset+int64(moovSize), 8+int64(len(mdat))); err != nil {
		return err
	}
	if err := mp4.WriteAtomAt(uint32(8+len(mdat)), "mdat", mdat, moovOffset+int64(moovSize)); err != nil {
		return err
	}

	return nil
}
