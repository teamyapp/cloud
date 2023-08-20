package entity

import (
	"io"
)

type File struct {
	Metadata     FileMetadata
	ChunksBuffer io.Reader
}
