package domain

import "context"

// ContentProvider reads documentation content from the filesystem.
type ContentProvider interface {
	// ListDocs returns all available documentation names.
	ListDocs(ctx context.Context) ([]string, error)

	// GetDoc returns metadata for a documentation.
	GetDoc(ctx context.Context, name string) (Doc, error)

	// GetDocSummary returns a lightweight public summary for a documentation.
	GetDocSummary(ctx context.Context, name string) (DocSummary, error)

	// GetPage returns a page's content and metadata.
	GetPage(ctx context.Context, doc string, page string) (Page, error)

	// SearchPages searches for query across all pages of a documentation.
	SearchPages(ctx context.Context, doc string, query string) ([]SearchResult, error)

	// GetRootConfig returns the root content configuration.
	GetRootConfig(ctx context.Context) (RootConfig, error)

	// GetLogoPath returns the filesystem path to the logo for a doc or the root site.
	GetLogoPath(ctx context.Context, doc string) (string, error)

	// GetAssetPath returns the filesystem path to a static asset inside a documentation directory.
	GetAssetPath(ctx context.Context, doc string, assetPath string) (string, error)

	// GetAssetAccess returns the effective access info for an asset path.
	GetAssetAccess(ctx context.Context, doc string, assetPath string) (AccessInfo, error)

	// GetUIText returns the UI text configuration for the frontend.
	GetUIText(ctx context.Context) (map[string]any, error)
}
