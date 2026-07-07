package domain

// PageNode represents a page or directory in the doc tree.
type PageNode struct {
	Path     string     `json:"path"`
	Title    string     `json:"title"`
	Access   string     `json:"access,omitempty"`
	HasIndex bool       `json:"has_index,omitempty"`
	Children []PageNode `json:"children,omitempty"`
}

// Doc represents a documentation collection.
type Doc struct {
	Name        string     `json:"name"`
	Title       string     `json:"title"`
	Description string     `json:"description,omitempty"`
	Theme       string     `json:"theme,omitempty"`
	Sidebar     []string   `json:"sidebar,omitempty"`
	Access      string     `json:"access,omitempty"`
	Pages       []PageNode `json:"pages,omitempty"`
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
