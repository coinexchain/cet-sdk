package msgqueue

import (
	"log"
	"os"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

// Prune backup-* file;
// Every file contains 10000 blocks data.

type PruneFile struct {
	doneHeightCh <-chan int
	dir          string
	ctx          sdk.Context
}

func NewPruneFile(ctx sdk.Context, heightCh <-chan int, dir string) *PruneFile {
	return &PruneFile{doneHeightCh: heightCh, dir: dir, ctx: ctx}
}

func (p *PruneFile) work() {
	for {
		doneHeight, ok := <-p.doneHeightCh
		if !ok {
			return
		}

		p.removeFiles(doneHeight)
	}
}

func (p *PruneFile) removeFiles(doneHeight int) {
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

func (p *PruneFile) timeToRemove(height int, doneHeight int) bool {
	currHeight := p.ctx.BlockHeight()
	_ = currHeight

	return false
}
