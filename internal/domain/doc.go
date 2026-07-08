// Package domain contains core business entities and errors for Vordoc.
package domain

// PageNode represents a page or directory in the doc tree.
type PageNode struct {
	Path        string     `json:"path"`
	Title       string     `json:"title"`
	Order       int        `json:"order,omitempty"`
	Access      string     `json:"access,omitempty"`
	AccessScope string     `json:"access_scope,omitempty"`
	HasIndex    bool       `json:"has_index,omitempty"`
	Show        bool       `json:"show,omitempty"`
	Children    []PageNode `json:"children,omitempty"`
}

// HeaderConfig represents header settings for a documentation or the root site.
type HeaderConfig struct {
	Enable   bool        `json:"enable"`
	Elements []string    `json:"elements,omitempty"`
	Title    string      `json:"title,omitempty"`
	Logo     *LogoConfig `json:"logo,omitempty"`
	Font     *FontConfig `json:"font,omitempty"`
}

// LogoConfig represents the logo settings for a header.
type LogoConfig struct {
	Path string `json:"path,omitempty"`
	Size int    `json:"size,omitempty"`
}

// FontConfig represents the font settings for a header.
type FontConfig struct {
	Name string `json:"name,omitempty"`
	Size int    `json:"size,omitempty"`
}

// ThemeConfig represents the site-wide theme selector settings.
type ThemeConfig struct {
	Default     string `json:"default,omitempty"`
	AccentColor string `json:"accent_color,omitempty"`
}

// RootPageConfig represents the root landing page settings.
type RootPageConfig struct {
	Enable bool   `json:"enable"`
	Title  string `json:"title,omitempty"`
}

// RootConfig represents the root content configuration.
type RootConfig struct {
	Root    RootPageConfig `json:"root"`
	Favicon string         `json:"favicon,omitempty"`
	Header  *HeaderConfig  `json:"header,omitempty"`
	Theme   *ThemeConfig   `json:"theme,omitempty"`
}

// Doc represents a documentation collection.
type Doc struct {
	Name         string        `json:"name"`
	Title        string        `json:"title"`
	Description  string        `json:"description,omitempty"`
	Access       string        `json:"access,omitempty"`
	AccessScope  string        `json:"access_scope,omitempty"`
	PasswordHash string        `json:"-"`
	Pages        []PageNode    `json:"pages,omitempty"`
	IndexPage    *Page         `json:"index_page,omitempty"`
	Header       *HeaderConfig `json:"header,omitempty"`
}

// Page represents a single documentation page.
type Page struct {
	Doc          string `json:"doc"`
	Path         string `json:"path"`
	FilePath     string `json:"filePath"`
	Title        string `json:"title"`
	Description  string `json:"description,omitempty"`
	Order        int    `json:"order,omitempty"`
	Content      string `json:"content,omitempty"`
	Access       string `json:"access,omitempty"`
	AccessScope  string `json:"-"`
	PasswordHash string `json:"-"` // not serialized, used internally
}

// SearchResult represents a single search hit within a documentation.
type SearchResult struct {
	Title       string `json:"title"`
	Path        string `json:"path"`
	Snippet     string `json:"snippet,omitempty"`
	Access      string `json:"access,omitempty"`
	AccessScope string `json:"-"`
	Score       int    `json:"-"` // not serialized, used for ranking
}

// AccessInfo describes the resolved access rule for a content node.
type AccessInfo struct {
	Access       string
	PasswordHash string
	Scope        string
}

// DocSummary is a lightweight public summary of a documentation.
type DocSummary struct {
	Name         string
	Title        string
	Description  string
	Access       string
	PasswordHash string
	Scope        string
}

// GlobalSearchPageResult is a single page hit within a documentation.
type GlobalSearchPageResult struct {
	Title   string `json:"title"`
	Path    string `json:"path"`
	Snippet string `json:"snippet,omitempty"`
}

// GlobalSearchResult groups search hits for one documentation.
type GlobalSearchResult struct {
	Name        string                   `json:"name"`
	Title       string                   `json:"title"`
	Description string                   `json:"description,omitempty"`
	Access      string                   `json:"access,omitempty"`
	Pages       []GlobalSearchPageResult `json:"pages,omitempty"`
}
