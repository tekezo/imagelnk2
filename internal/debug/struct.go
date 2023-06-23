package debug

import "imagelnk2/internal/core"

type (
	Entry struct {
		Filename string      `json:"filename"`
		URL      string      `json:"url"`
		Result   core.Result `json:"result"`
	}
)
