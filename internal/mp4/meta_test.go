package mp4

import (
	"bytes"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestMetadataWrite(t *testing.T) {
	metadata := Metadata{
		Title:       "A very long title for testing",
		Artist:      "Some artist",
		Album:       "Album",
		Description: "Some description",
		Copyright:   "2024",
		Released:    time.Date(2024, 11, 9, 12, 29, 56, 0, time.UTC),
	}

	target := filepath.Join(t.TempDir(), "with-metadata.m4a")
	require.NoError(t, copyFile("./empty.m4a", target))

	file, err := os.OpenFile(target, os.O_RDWR, 0)
	require.NoError(t, err)
	defer file.Close()

	require.NoError(t, metadata.Write(file))

	var output bytes.Buffer

	ffprobe := exec.Command("ffprobe", target)
	ffprobe.Stderr = &output
	ffprobe.Stdout = &output
	require.NoError(t, ffprobe.Run())

	metadataRegexp := regexp.MustCompile(`(?m)  Metadata:\n(    \w+ *: *[^\n]+\n)*`)
	whitespaceRegexp := regexp.MustCompile(`(?m)^ +| +$`)
	createdMetadata := metadataRegexp.FindString(output.String())
	createdMetadata = whitespaceRegexp.ReplaceAllString(createdMetadata, "")

	expectedMetadata := `Metadata:
major_brand     : M4A
minor_version   : 512
compatible_brands: M4A isomiso2
title           : A very long title for testing
artist          : Some artist
album           : Album
description     : Some description
copyright       : 2024
date            : 2024-11-09T12:29:56Z
`

	assert.Equal(t, expectedMetadata, createdMetadata)
}

func copyFile(from string, to string) error {
	existing, err := os.Open(from)
	if err != nil {
		return err
	}
	defer existing.Close()

	stat, err := existing.Stat()
	if err != nil {
		return err
	}

	copy, err := os.OpenFile(to, os.O_CREATE|os.O_WRONLY, stat.Mode())
	if err != nil {
		return err
	}
	defer copy.Close()

	_, err = io.Copy(copy, existing)
	return err
}
