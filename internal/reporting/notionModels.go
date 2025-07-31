package reporting

type NotionPagesRoot struct {
	Object  string `json:"object"`
	Results []struct {
		Object         string `json:"object"`
		ID             string `json:"id"`
		Parent         Parent `json:"parent"`
		CreatedTime    string `json:"created_time"`
		LastEditedTime string `json:"last_edited_time"`
		CreatedBy      User   `json:"created_by"`
		LastEditedBy   User   `json:"last_edited_by"`
		HasChildren    bool   `json:"has_children"`
		Archived       bool   `json:"archived"`
		InTrash        bool   `json:"in_trash"`
		Type           string `json:"type"`
		ChildPage      struct {
			Title string `json:"title"`
		} `json:"child_page"`
	} `json:"results"`
	NextCursor string `json:"next_cursor"`
	HasMore    bool   `json:"has_more"`
	RequestID  string `json:"request_id"`
}

type Parent struct {
	Type   string `json:"type"`
	PageID string `json:"page_id"`
}

type User struct {
	Object string `json:"object"`
	ID     string `json:"id"`
}

type PageRequest struct {
	Parent struct {
		Type   string `json:"type"`
		PageID string `json:"page_id"`
	} `json:"parent"`
	Properties struct {
		Title struct {
			ID    string `json:"id"`
			Type  string `json:"type"`
			Title []struct {
				Type string `json:"type"`
				Text struct {
					Content string `json:"content"`
				} `json:"text"`
			} `json:"title"`
		} `json:"title"`
	} `json:"properties"`
}

type ToggleBlock struct {
	Object string `json:"object"`
	Type   string `json:"type"`
	Toggle struct {
		RichText []struct {
			Type string `json:"type"`
			Text struct {
				Content string `json:"content"`
				Link    any    `json:"link"`
			} `json:"text"`
			Annotations struct {
				Underline bool   `json:"underline"`
				Bold      bool   `json:"bold"`
				Color     string `json:"color"`
			} `json:"annotations"`
		} `json:"rich_text"`
		Children []struct {
			Object    string `json:"object"`
			Type      string `json:"type"`
			Paragraph struct {
				RichText []struct {
					Type string `json:"type"`
					Text struct {
						Content string `json:"content"`
					} `json:"text"`
				} `json:"rich_text"`
			} `json:"paragraph"`
		} `json:"children"`
	} `json:"toggle"`
}
