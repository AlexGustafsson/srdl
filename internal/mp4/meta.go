package mp4

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"reflect"
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
//
// Audiobookshelf:
//
// See: https://www.audiobookshelf.org/docs#book-audio-metadata
//
//   - Author
//   - Title
//   - Subtitle
//   - Publisher
//   - Published year
//   - ...
type Metadata struct {
	Title       string    `atom:"\xa9nam"`
	Artist      string    `atom:"\xa9ART"`
	Album       string    `atom:"\xa9alb"`
	Description string    `atom:"desc"`
	Copyright   string    `atom:"\xa9cpy"`
	Released    time.Time `atom:"\xa9day"`
}

// Bytes returns the MP4 byte representation of the metadata, to be put into a
// ilst atom.
func (m Metadata) Bytes() []byte {
	var buffer bytes.Buffer

	structValue := reflect.ValueOf(m)
	structType := structValue.Type()
	for i := 0; i < structType.NumField(); i++ {
		fieldType := structType.Field(i)

		if !fieldType.IsExported() {
			continue
		}

		atomType := fieldType.Tag.Get("atom")
		if atomType == "" {
			continue
		}

		fieldValue := structValue.Field(i)
		if fieldValue.IsZero() {
			continue
		}

		formattedValue := ""
		switch v := fieldValue.Interface().(type) {
		case string:
			formattedValue = v
		case time.Time:
			formattedValue = v.Format(time.RFC3339)
		default:
			panic(fmt.Errorf("invalid metadata field of type %s", fieldType.Type.String()))
		}

		// Write the atom header, the atom will contain a data atom
		if err := writeAtomHeader(&buffer, uint32(8+8+8+len(formattedValue)), atomType); err != nil {
			panic(err)
		}

		// Write the data header, the atom will contain the value of the field
		if err := writeAtomHeader(&buffer, uint32(8+8+len(formattedValue)), "data"); err != nil {
			panic(err)
		}

		// NOTE: the bytes being written are "the_type" | "the_locale".
		// FFMPEG never seems to set these to anything else than the following
		// bytes.
		// SEE: https://developer.apple.com/documentation/quicktime-file-format/metadata_item_list_atom
		if _, err := buffer.Write([]byte{0x00, 0x00, 0x00, 0x01, 0x00, 0x00, 0x00, 0x00}); err != nil {
			panic(err)
		}

		// Write the actual value
		if _, err := buffer.WriteString(formattedValue); err != nil {
			panic(err)
		}
	}

	return buffer.Bytes()
}

func (m Metadata) Write(f *os.File) error {
	mp4 := File{f}

	stat, err := mp4.Stat()
	if err != nil {
		return err
	}

	moovOffset, moovSize, err := mp4.SeekAtom("moov")
	if err != nil {
		return err
	}

	udtaOffset, udtaSize, err := mp4.SeekAtom("udta")
	if err != nil {
		return err
	}

	metaOffset, metaSize, err := mp4.SeekAtom("meta")
	if err != nil {
		return err
	}

	// The meta atom has a larger header, skip it
	mp4.Seek(4, io.SeekCurrent)

	ilstOffset, ilstSize, err := mp4.SeekAtom("ilst")
	if err != nil {
		return err
	}

	mb := m.Bytes()

	dsize := len(mb) + 8 - int(ilstSize)

	if _, err := f.WriteAt(formatAtomHeader(uint32(int(moovSize)+dsize), "moov"), moovOffset); err != nil {
		return err
	}

	if _, err := f.WriteAt(formatAtomHeader(uint32(int(udtaSize)+dsize), "udta"), udtaOffset); err != nil {
		return err
	}

	if _, err := f.WriteAt(formatAtomHeader(uint32(int(metaSize)+dsize), "meta"), metaOffset); err != nil {
		return err
	}

	if _, err := f.WriteAt(formatAtomHeader(uint32(int(ilstSize)+dsize), "ilst"), ilstOffset); err != nil {
		return err
	}

	if dsize > 0 {
		if err := f.Truncate(stat.Size() + int64(dsize)); err != nil {
			return err
		}
	}

	if _, err := f.WriteAt(mb, ilstOffset+HeaderSize); err != nil {
		return err
	}

	// Assumes the ilst atom is the last atom as nothing will be written after it

	return nil
}
