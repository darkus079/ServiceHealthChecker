package models

type CheckRequest struct {
	Links []string `json:"links"`
}

type CheckResponse struct {
	Links    map[string]string `json:"links"`
	LinksNum int               `json:"links_num"`
}

type ReportRequest struct {
	LinksList []int `json:"links_list"`
}

type LinkStatus struct {
	URL    string `json:"url"`
	Status string `json:"status"`
}

type LinkSet struct {
	ID       int          `json:"id"`
	Links    []LinkStatus `json:"links"`
	Checked  bool         `json:"checked"`
}

type StorageData struct {
	LastID   int       `json:"last_id"`
	LinkSets []LinkSet `json:"link_sets"`
}

type PendingTask struct {
	ID    int      `json:"id"`
	Links []string `json:"links"`
}

type PendingTasks struct {
	Tasks []PendingTask `json:"tasks"`
}

