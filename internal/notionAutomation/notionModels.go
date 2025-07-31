package notionAutomation

const (
	NotionReportingUrl      = "https://api.notion.com/v1/pages"
	MainNotionReportingPage = "200e8162bffd49fbadfb6764bf5c06ff"
)

type NotionReportingPayloadRoot struct {
	Parent     Parent     `json:"parent"`
	Properties Properties `json:"properties"`
	Children   []Child    `json:"children"`
}

type Parent struct {
	Type   string `json:"type"`
	PageId string `json:"page_id"`
}

type Properties struct {
	Title Title `json:"title"`
}

type Child struct {
	Object    string    `json:"object"`
	Type      string    `json:"type"`
	Paragraph Paragraph `json:"paragraph"`
}

type Title struct {
	Id    string         `json:"id"`
	Type  string         `json:"type"`
	Title []TitleContent `json:"title"`
}

type TitleContent struct {
	Type string `json:"type"`
	Text Text   `json:"text"`
}

type Text struct {
	Content string `json:"content"`
	Link    string `json:"link"`
}

type Paragraph struct {
	RichText []RichText `json:"text"`
}

type RichText struct {
	Type        string      `json:"type"`
	Text        Text        `json:"text"`
	Annotations Annotations `json:"annotations"`
}

type Annotations struct {
	Bold  bool   `json:"bold"`
	Color string `json:"color"`
}

type NotionPage struct {
	Object  string `json:"object"`
	Results []Page `json:"results"`
}

type Page struct {
	Object         string    `json:"object"`
	Id             string    `json:"id"`
	Parent         Parent    `json:"parent"`
	CreatedTime    string    `json:"created_time"`
	LastEditedTime User      `json:"last_edited_time"`
	HasChildren    bool      `json:"has_children"`
	Archived       bool      `json:"archived"`
	InTrash        bool      `json:"in_trash"`
	Type           string    `json:"type"`
	ChildPage      ChildPage `json:"child_page"`
}

type User struct {
	Object string `json:"object"`
	Id     string `json:"id"`
}

type ChildPage struct {
	Title string `json:"title"`
}
