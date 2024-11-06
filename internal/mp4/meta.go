package mp4

import (
	"bytes"
	"encoding/binary"
	"io"
	"os"
	"time"
)

// Metadata contains common metadata fields.
//
// Jellyfin:
//
// See: https://github.com/jellyfin/jellyfin/blob/1dd3792984416e5ff365cd259b270eab94c0cd5a/MediaBrowser.Providers/MediaInfo/FFProbeVideoInfo.cs#L31
// See: https://github.com/jellyfin/jellyfin/blob/bf00899f92881c987f019ad7d20f0cef42d4e3e7/MediaBrowser.MediaEncoding/Probing/ProbeResultNormalizer.cs#L1212
//
//   - Name
//   - Overview
//   - Cast
//   - Official rating
//   - Genres
//   - Studios
//
// Kodi:
//
// See: https://kodi.wiki/view/Video_file_tagging
//
//   - Album
//   - Artist
//   - Writing Credits
//   - Year
//   - Genre
//   - Plot
//   - Plot Outline
//   - Title
//   - Track
type Metadata struct {
	Title       string    `box:"\xa9nam"`
	Artist      string    `box:"\xa9ART"`
	Album       string    `box:"\xa9alb"`
	Description string    `box:"desc"`
	Copyright   string    `box:"\xa9cpy"`
	Released    time.Time `box:"\xa9day"`
}

func (m Metadata) Bytes() []byte {
	var buffer bytes.Buffer

	// TODO: Rewrite
	// NOTE: the bytes being written are "the_type" | "the_locale":
	// https://developer.apple.com/documentation/quicktime-file-format/metadata_item_list_atom

	if m.Title != "" {
		if err := writeBoxHeader(&buffer, uint32(8+8+8+len(m.Title)), "\xa9nam"); err != nil {
			panic(err)
		}
		if err := writeBoxHeader(&buffer, uint32(8+8+len(m.Title)), "data"); err != nil {
			panic(err)
		}
		if _, err := buffer.Write([]byte{0x00, 0x00, 0x00, 0x01, 0x00, 0x00, 0x00, 0x00}); err != nil {
			panic(err)
		}
		if _, err := buffer.WriteString(m.Title); err != nil {
			panic(err)
		}
	}

	if m.Artist != "" {
		if err := writeBoxHeader(&buffer, uint32(8+8+8+len(m.Artist)), "\xa9ART"); err != nil {
			panic(err)
		}
		if err := writeBoxHeader(&buffer, uint32(8+8+len(m.Artist)), "data"); err != nil {
			panic(err)
		}
		if _, err := buffer.Write([]byte{0x00, 0x00, 0x00, 0x01, 0x00, 0x00, 0x00, 0x00}); err != nil {
			panic(err)
		}
		if _, err := buffer.WriteString(m.Artist); err != nil {
			panic(err)
		}
	}

	if m.Album != "" {
		if err := writeBoxHeader(&buffer, uint32(8+8+8+len(m.Album)), "\xa9alb"); err != nil {
			panic(err)
		}
		if err := writeBoxHeader(&buffer, uint32(8+8+len(m.Album)), "data"); err != nil {
			panic(err)
		}
		if _, err := buffer.Write([]byte{0x00, 0x00, 0x00, 0x01, 0x00, 0x00, 0x00, 0x00}); err != nil {
			panic(err)
		}
		if _, err := buffer.WriteString(m.Album); err != nil {
			panic(err)
		}
	}

	if m.Description != "" {
		if err := writeBoxHeader(&buffer, uint32(8+8+8+len(m.Description)), "desc"); err != nil {
			panic(err)
		}
		if err := writeBoxHeader(&buffer, uint32(8+8+len(m.Description)), "data"); err != nil {
			panic(err)
		}
		if _, err := buffer.Write([]byte{0x00, 0x00, 0x00, 0x01, 0x00, 0x00, 0x00, 0x00}); err != nil {
			panic(err)
		}
		if _, err := buffer.WriteString(m.Description); err != nil {
			panic(err)
		}
	}

	if m.Copyright != "" {
		if err := writeBoxHeader(&buffer, uint32(8+8+8+len(m.Copyright)), "\xa9cpy"); err != nil {
			panic(err)
		}
		if err := writeBoxHeader(&buffer, uint32(8+8+len(m.Copyright)), "data"); err != nil {
			panic(err)
		}
		if _, err := buffer.Write([]byte{0x00, 0x00, 0x00, 0x01, 0x00, 0x00, 0x00, 0x00}); err != nil {
			panic(err)
		}
		if _, err := buffer.WriteString(m.Copyright); err != nil {
			panic(err)
		}
	}

	if !m.Released.IsZero() {
		// Released: Format 2013-07-20T08:03:13+1000 or 2022:08:28 15:25:09?
		b := m.Released.UTC().Format(time.RFC3339)

		if err := writeBoxHeader(&buffer, uint32(8+8+8+len(b)), "\xa9day"); err != nil {
			panic(err)
		}
		if err := writeBoxHeader(&buffer, uint32(8+8+len(b)), "data"); err != nil {
			panic(err)
		}
		if _, err := buffer.Write([]byte{0x00, 0x00, 0x00, 0x01, 0x00, 0x00, 0x00, 0x00}); err != nil {
			panic(err)
		}
		if _, err := buffer.WriteString(b); err != nil {
			panic(err)
		}
	}

	return buffer.Bytes()
}

func (m Metadata) Write(f *os.File) error {
	stat, err := f.Stat()
	if err != nil {
		return err
	}

	moovOffset, moovSize, err := seekBox(f, "moov")
	if err != nil {
		return err
	}

	udtaOffset, udtaSize, err := seekBox(f, "udta")
	if err != nil {
		return err
	}

	metaOffset, metaSize, err := seekBox(f, "meta")
	if err != nil {
		return err
	}

	// The meta box has a larger header, skip it
	f.Seek(4, io.SeekCurrent)

	ilstOffset, ilstSize, err := seekBox(f, "ilst")
	if err != nil {
		return err
	}

	mb := m.Bytes()

	dsize := len(mb) + 8 - int(ilstSize)

	if _, err := f.WriteAt(formatBoxHeader(uint32(int(moovSize)+dsize), "moov"), moovOffset); err != nil {
		return err
	}

	if _, err := f.WriteAt(formatBoxHeader(uint32(int(udtaSize)+dsize), "udta"), moovOffset+8+udtaOffset); err != nil {
		return err
	}

	if _, err := f.WriteAt(formatBoxHeader(uint32(int(metaSize)+dsize), "meta"), moovOffset+8+udtaOffset+8+metaOffset); err != nil {
		return err
	}

	if _, err := f.WriteAt(formatBoxHeader(uint32(int(ilstSize)+dsize), "ilst"), moovOffset+8+udtaOffset+8+metaOffset+12+ilstOffset); err != nil {
		return err
	}

	if dsize > 0 {
		if err := f.Truncate(stat.Size() + int64(dsize)); err != nil {
			return err
		}
	}

	if _, err := f.WriteAt(mb, moovOffset+8+udtaOffset+8+metaOffset+12+ilstOffset+8); err != nil {
		return err
	}

	// Assumes the ilst box is the last box as nothing will be written after it

	return nil
}

func seekBox(r io.ReadSeeker, needle string) (int64, uint32, error) {
	var boxOffset uint32 = 0
	for {
		boxSize, boxType, err := readBoxHeader(r)
		if err == io.EOF {
			return -1, 0, nil
		} else if err != nil {
			return -1, 0, err
		}

		if boxType == needle {
			return int64(boxOffset), boxSize, nil
		}

		// Seek past the box's data
		if _, err := r.Seek(int64(boxSize)-8, io.SeekCurrent); err != nil {
			return -1, 0, err
		}

		boxOffset += boxSize
	}
}

func readBoxHeader(r io.Reader) (uint32, string, error) {
	var buffer [8]byte
	if _, err := r.Read(buffer[:]); err != nil {
		return 0, "", err
	}

	return binary.BigEndian.Uint32(buffer[0:4]), string(buffer[4:8]), nil
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
