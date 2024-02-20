package task

import (
	"errors"
	"path/filepath"
	"strings"

	"github.com/btcsuite/btcutil/base58"
	"github.com/google/uuid"
)

type logfile struct {
	module string
	id     uuid.UUID
}

func newLogfileFromPath(path string) (logfile, error) {
	path, found := strings.CutPrefix(path, ".pug")
	if !found {
		return logfile{}, errors.New("bad")
	}
	base := filepath.Base(path)
	base58id, found := strings.CutSuffix(base, ".log")
	if !found {
		return logfile{}, errors.New("bad")
	}
	id, err := uuid.FromBytes(base58.Decode(base58id))
	if err != nil {
		return logfile{}, errors.New("bad")
	}
	return logfile{
		module: filepath.Dir(path),
		id:     id,
	}, nil
}

func (l logfile) String() string {
	return filepath.Join(".pug", l.module, l.id.String()+".log")
}
