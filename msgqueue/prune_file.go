package msgqueue

import (
	"log"
	"os"
)

type ExpectGetHeight interface {
	GetHeight() int64
}

// Prune backup-* file;
// Every file contains 10000 blocks data.

type PruneFile struct {
	doneHeightCh <-chan int64
	dir          string
	ExpectGetHeight
}

func NewPruneFile(ght ExpectGetHeight, heightCh <-chan int64, dir string) *PruneFile {
	return &PruneFile{doneHeightCh: heightCh, dir: dir, ExpectGetHeight: ght}
}

func (p *PruneFile) Work() {
	for {
		doneHeight, ok := <-p.doneHeightCh
		if !ok {
			return
		}

		p.removeFiles(doneHeight)
	}
}

func (p *PruneFile) removeFiles(doneHeight int64) {
	fileName, leastHeight := p.getLeastHeightFileFromDir()
	if p.timeToRemove(leastHeight, doneHeight) {
		if err := os.Remove(fileName); err != nil {
			log.Fatalln(err)
		}
	}
}

func (p *PruneFile) getLeastHeightFileFromDir() (fileName string, leastHeight int) {

	return
}

func (p *PruneFile) timeToRemove(height int, doneHeight int64) bool {
	currHeight := p.GetHeight()
	_ = currHeight

	return false
}
