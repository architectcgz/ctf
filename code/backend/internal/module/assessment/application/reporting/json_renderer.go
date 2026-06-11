package reporting

import (
	"encoding/json"
	"os"

	"ctf-platform/internal/apperror"
)

func WriteJSONReport(filePath string, data any) error {
	content, err := json.MarshalIndent(data, "", "  ")
	if err != nil {
		return apperror.ErrInternal.WithCause(err)
	}
	content = append(content, '\n')
	if err := os.WriteFile(filePath, content, 0o644); err != nil {
		return apperror.ErrInternal.WithCause(err)
	}
	return nil
}
