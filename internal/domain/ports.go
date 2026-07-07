package domain

import "context"

// ContentProvider reads documentation content from the filesystem.
type ContentProvider interface {
	// ListDocs returns all available documentation names.
	ListDocs(ctx context.Context) ([]string, error)

	// GetDoc returns metadata for a documentation.
	GetDoc(ctx context.Context, name string) (Doc, error)

	// GetPage returns a page's content and metadata.
	GetPage(ctx context.Context, doc string, page string) (Page, error)

	// GetTheme returns the CSS variables for a theme.
	GetTheme(ctx context.Context, name string) (Theme, error)

	// GetRootConfig returns the root content configuration.
	GetRootConfig(ctx context.Context) (RootConfig, error)

	// GetLogoPath returns the filesystem path to the logo for a doc or the root site.
	GetLogoPath(ctx context.Context, doc string) (string, error)
}
