package bulk

// Row is a parsed data row.
type Row struct {
	Index  int
	Fields map[string]string
}

// RowError is an error for a specific row.
type RowError struct {
	Row   int    `json:"row"`
	Error string `json:"error"`
}

// BulkResult is a summary of bulk processing.
type BulkResult struct {
	Total     int        `json:"total"`
	Succeeded int        `json:"succeeded"`
	Failed    int        `json:"failed"`
	Errors    []RowError `json:"errors,omitempty"`
}
