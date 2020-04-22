package msgqueue

import (
	"fmt"
	"os"
)

// Prune backup-* file;
// Every file contains 10000 blocks data.

type FileDeleter struct {
	doneHeightCh <-chan int64
	dir          string
}

func NewFileDeleter(heightCh <-chan int64, dir string) *FileDeleter {
	return &FileDeleter{doneHeightCh: heightCh, dir: dir}
}

func (p *FileDeleter) Run() {
	go func() {
		for {
			doneHeight, ok := <-p.doneHeightCh
			if !ok {
				return
			}

			p.removeFiles(doneHeight)
		}
	}()
}

func (p *FileDeleter) removeFiles(doneHeight int64) {
	fileName, leastHeight := p.getLeastHeightFileFromDir()
	if p.timeToRemove(leastHeight, doneHeight) {
		if err := os.Remove(fileName); err != nil {
			panic(fmt.Sprintf("Remove file from dir failed; dir[%s], file[%s]\n", p.dir, fileName))
		}
	}
}

func (p *FileDeleter) getLeastHeightFileFromDir() (fileName string, leastHeight int64) {
	fileName, height, err := GetFileLeastHeightInDir(p.dir)
	if err != nil {
		panic(fmt.Sprintf("Get block Height from files failed; dir[%s], error[%s]\n", p.dir, err.Error()))
	}
	return fileName, height
}

func (p *FileDeleter) timeToRemove(leastHeight int64, doneHeight int64) bool {
	return doneHeight-leastHeight >= int64(2*FILEHEIGHT)
}
