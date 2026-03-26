package domain

// Batch represents result row from DAX batch lookup.
type Batch struct {
	Batch       string `json:"batch"`
	NameAlias   string `json:"namealias"`
	ConfigID    string `json:"configid,omitempty"`
	ColorID     string `json:"colorid,omitempty"`
	WmsLocation string `json:"wmslocation"`
	License     string `json:"license"`
	UserName    string `json:"username,omitempty"`
}
