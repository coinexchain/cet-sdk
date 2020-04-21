package msgqueue

import (
	"fmt"
	"os"
)

// Prune backup-* file;
// Every file contains 10000 blocks data.

type PruneFile struct {
	doneHeightCh <-chan int64
	dir          string
}

func NewPruneFile(heightCh <-chan int64, dir string) *PruneFile {
	return &PruneFile{doneHeightCh: heightCh, dir: dir}
}

func (p *PruneFile) Work() {
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

func (p *PruneFile) removeFiles(doneHeight int64) {
	fileName, leastHeight := p.getLeastHeightFileFromDir()
	if p.timeToRemove(leastHeight, doneHeight) {
		if err := os.Remove(fileName); err != nil {
			panic(fmt.Sprintf("Remove file from dir failed; dir[%s], file[%s]\n", p.dir, fileName))
		}
	}
}

func (p *PruneFile) getLeastHeightFileFromDir() (fileName string, leastHeight int64) {
	fileName, height, err := GetFileLeastHeightInDir(p.dir)
	if err != nil {
		panic(fmt.Sprintf("Get block Height from files failed; dir[%s], error[%s]\n", p.dir, err.Error()))
	}
	return fileName, height
}

func (p *PruneFile) timeToRemove(leastHeight int64, doneHeight int64) bool {
	return doneHeight-leastHeight > int64(2*FILEHEIGHT)
}
