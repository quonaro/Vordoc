package domain

// PageNode represents a page or directory in the doc tree.
type PageNode struct {
	Path     string     `json:"path"`
	Title    string     `json:"title"`
	Access   string     `json:"access,omitempty"`
	HasIndex bool       `json:"has_index,omitempty"`
	Children []PageNode `json:"children,omitempty"`
}

// HeaderConfig represents header settings for a documentation or the root site.
type HeaderConfig struct {
	Enable bool   `json:"enable"`
	Title  string `json:"title,omitempty"`
	Logo   string `json:"logo,omitempty"`
}

// RootConfig represents the root content configuration.
type RootConfig struct {
	EnableRootPage bool          `json:"enable_root_page"`
	Header         *HeaderConfig `json:"header,omitempty"`
}

// Doc represents a documentation collection.
type Doc struct {
	Name        string        `json:"name"`
	Title       string        `json:"title"`
	Description string        `json:"description,omitempty"`
	Theme       string        `json:"theme,omitempty"`
	Sidebar     []string      `json:"sidebar,omitempty"`
	Access      string        `json:"access,omitempty"`
	Pages       []PageNode    `json:"pages,omitempty"`
	IndexPage   *Page         `json:"index_page,omitempty"`
	Header      *HeaderConfig `json:"header,omitempty"`
}

// Page represents a single documentation page.
type Page struct {
	Doc          string `json:"doc"`
	Path         string `json:"path"`
	Title        string `json:"title"`
	Order        int    `json:"order,omitempty"`
	Content      string `json:"content,omitempty"`
	Access       string `json:"access,omitempty"`
	PasswordHash string `json:"-"` // not serialized, used internally
}

// Theme represents a theme definition.
type Theme struct {
	Name    string `json:"name"`
	VarsCSS string `json:"vars_css,omitempty"`
}
