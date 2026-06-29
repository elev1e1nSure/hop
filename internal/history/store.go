package history

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"os"

	"hop/internal/apperr"
	"hop/internal/domain"
	"hop/internal/util"
)

type file struct {
	Servers map[string]domain.HistoryRecord `json:"servers"`
}

func Load(path string) (map[string]domain.HistoryRecord, error) {
	data, err := os.ReadFile(path)
	if errors.Is(err, os.ErrNotExist) {
		return map[string]domain.HistoryRecord{}, nil
	}
	if err != nil {
		cause := fmt.Errorf("read history file %q: %w", path, err)
		return nil, apperr.Wrap(apperr.ErrReadHistory, cause, path)
	}
	if len(bytes.TrimSpace(data)) == 0 {
		return map[string]domain.HistoryRecord{}, nil
	}

	var stored file
	if err := json.Unmarshal(data, &stored); err != nil {
		cause := fmt.Errorf("parse history file %q: %w", path, err)
		return nil, apperr.Wrap(apperr.ErrParseHistory, cause, path)
	}
	if stored.Servers == nil {
		stored.Servers = map[string]domain.HistoryRecord{}
	}
	return stored.Servers, nil
}

func Save(path string, records map[string]domain.HistoryRecord) error {
	if records == nil {
		records = map[string]domain.HistoryRecord{}
	}
	data, err := json.MarshalIndent(file{Servers: records}, "", "  ")
	if err != nil {
		cause := fmt.Errorf("marshal history for %q: %w", path, err)
		return apperr.Wrap(apperr.ErrMarshalHistory, cause, path)
	}
	data = append(data, '\n')
	if err := util.AtomicWrite(path, data, 0o600, 0o700); err != nil {
		cause := fmt.Errorf("write history file %q: %w", path, err)
		return apperr.Wrap(apperr.ErrWriteHistory, cause, path)
	}
	return nil
}
