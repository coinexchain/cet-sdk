package msgqueue

import (
	"bytes"
	"fmt"
	"io"
)

const (
	filePrefix = "backup-"
)

var MaxFileSize = 1024 * 1024 * 100

var _ MsgWriter = (*dirMsgWriter)(nil)

type GetFilePathAndFileIndexFromDirCb func(string, int) (filePath string, fileIndex int, err error)

type dirMsgWriter struct {
	io.WriteCloser
	haveWriteSize int
	fileIndex     int
	dir           string
	timeNewFile   func(k, v []byte) bool
}

func NewDirMsgWriter(dir string, cb GetFilePathAndFileIndexFromDirCb) (MsgWriter, error) {
	filePath, fileIndex, err := cb(dir, MaxFileSize)
	if err != nil {
		return &dirMsgWriter{}, err
	}
	file, err := openFile(filePath)
	if err != nil {
		return &dirMsgWriter{}, err
	}
	fileSize := GetFileSize(filePath)
	if fileSize < 0 {
		return &dirMsgWriter{}, fmt.Errorf("The parameter passed in is not the correct file path. ")
	}
	diw := &dirMsgWriter{
		WriteCloser:   file,
		fileIndex:     fileIndex,
		dir:           dir,
		haveWriteSize: fileSize,
	}
	diw.timeNewFile = diw.timeToNewFile()
	return diw, nil
}

func (w *dirMsgWriter) WriteKV(k, v []byte) error {
	if w.timeNewFile(k, v) {
		if err := w.Close(); err != nil {
			return err
		}
		file, err := openFile(GetFileName(w.dir, w.fileIndex+1))
		if err != nil {
			return err
		}
		w.WriteCloser = file
		w.fileIndex++
		w.haveWriteSize = 0
	}
	buferr := bytes.NewBuffer(nil)
	buferr.Write(k)
	buferr.Write([]byte("#"))
	buferr.Write(v)
	buferr.Write([]byte("\r\n"))
	if _, err := w.WriteCloser.Write(buferr.Bytes()); err != nil {
		return err
	}
	w.haveWriteSize += len(k) + len(v) + 3
	return nil
}

func (w *dirMsgWriter) Close() error {
	return w.WriteCloser.Close()
}

func (w *dirMsgWriter) String() string {
	return "dir"
}

func (w *dirMsgWriter) SetTimeToNewFile(cb func(k, v []byte) bool) {
	w.timeNewFile = cb
}

func (w *dirMsgWriter) timeToNewFile() func(k, v []byte) bool {
	return func(k, v []byte) bool {
		return len(k)+len(v)+3+w.haveWriteSize > MaxFileSize
	}
}
