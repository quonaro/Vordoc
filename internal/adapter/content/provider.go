package content

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"vordoc/internal/domain"
)

// Provider implements domain.ContentProvider by reading from the filesystem.
type Provider struct {
	root   string
	logger *slog.Logger
}

// NewProvider creates a filesystem content provider.
func NewProvider(root string, logger *slog.Logger) *Provider {
	return &Provider{
		root:   root,
		logger: logger,
	}
}

// ListDocs returns all documentation directory names.
func (p *Provider) ListDocs(_ context.Context) ([]string, error) {
	entries, err := os.ReadDir(p.root)
	if err != nil {
		return nil, fmt.Errorf("reading content root: %w", err)
	}

	var docs []string
	for _, e := range entries {
		if !e.IsDir() {
			continue
		}
		if strings.HasPrefix(e.Name(), ".") {
			continue
		}
		configPath := filepath.Join(p.root, e.Name(), "config.yaml")
		if _, err := os.Stat(configPath); err == nil {
			docs = append(docs, e.Name())
		}
	}

	return docs, nil
}

// GetDoc returns metadata for a documentation.
func (p *Provider) GetDoc(ctx context.Context, name string) (domain.Doc, error) {
	docPath := filepath.Join(p.root, name)
	info, err := os.Stat(docPath)
	if err != nil {
		if os.IsNotExist(err) {
			return domain.Doc{}, fmt.Errorf("%w: %s", domain.ErrDocNotFound, name)
		}
		return domain.Doc{}, fmt.Errorf("stat doc: %w", err)
	}
	if !info.IsDir() {
		return domain.Doc{}, fmt.Errorf("%w: %s is not a directory", domain.ErrDocNotFound, name)
	}

	cfg, err := loadDocConfig(filepath.Join(docPath, "config.yaml"))
	if err != nil {
		return domain.Doc{}, err
	}

	doc := domain.Doc{
		Name:   name,
		Title:  cfg.Title,
		Header: p.resolveDocHeader(name, cfg),
	}

	pages, _ := p.scanDocPages(docPath)
	doc.Pages = pages
	doc.Access = p.docAccess(docPath)

	// Load root index page if present.
	if idx, err := p.GetPage(ctx, name, ""); err == nil {
		doc.IndexPage = &idx
		doc.Description = idx.Description
		if doc.Title == "" && idx.Title != "" {
			doc.Title = idx.Title
		}
	}

	if doc.Title == "" {
		doc.Title = name
	}

	return doc, nil
}

// docAccess returns the access level for the documentation root index page.
func (p *Provider) docAccess(docPath string) string {
	idx := filepath.Join(docPath, "index.md")
	var fm map[string]any
	if data, err := os.ReadFile(idx); err == nil {
		fm, _, _ = parseFrontmatter(data)
	}
	access, _ := resolveAccess(docPath, idx, fm)
	return access
}

func (p *Provider) scanDocPages(docPath string) ([]domain.PageNode, error) {
	return p.scanDir(docPath, docPath)
}

func (p *Provider) scanDir(dir string, docPath string) ([]domain.PageNode, error) {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return nil, err
	}

	var nodes []domain.PageNode
	for _, e := range entries {
		name := e.Name()
		if strings.HasPrefix(name, ".") {
			continue
		}

		fullPath := filepath.Join(dir, name)
		rel, _ := filepath.Rel(docPath, fullPath)
		rel = filepath.ToSlash(rel)

		if e.IsDir() {
			children, err := p.scanDir(fullPath, docPath)
			if err != nil {
				return nil, err
			}
			node := domain.PageNode{
				Path:     rel,
				Title:    name,
				Children: children,
			}
			hasIndex := false
			idx := filepath.Join(fullPath, "index.md")
			if data, err := os.ReadFile(idx); err == nil {
				hasIndex = true
				fm, _, _ := parseFrontmatter(data)
				if t := getString(fm, "title", ""); t != "" {
					node.Title = t
				}
				node.Order = getInt(fm, "order", 0)
				node.Access, _ = resolveAccess(docPath, idx, fm)
			}
			if len(children) == 0 && !hasIndex {
				continue
			}
			node.HasIndex = hasIndex
			nodes = append(nodes, node)
		} else if filepath.Ext(name) == ".md" {
			if name == "index.md" && dir == docPath {
				continue
			}
			if name == "index.md" {
				continue
			}

			data, err := os.ReadFile(fullPath)
			if err != nil {
				continue
			}
			fm, _, _ := parseFrontmatter(data)
			title := getString(fm, "title", strings.TrimSuffix(name, ".md"))
			order := getInt(fm, "order", 0)
			access, _ := resolveAccess(docPath, fullPath, fm)
			nodes = append(nodes, domain.PageNode{
				Path:   strings.TrimSuffix(rel, ".md"),
				Title:  title,
				Order:  order,
				Access: access,
			})
		}
	}

	sort.Slice(nodes, func(i, j int) bool {
		if nodes[i].Order != nodes[j].Order {
			return nodes[i].Order < nodes[j].Order
		}
		return nodes[i].Path < nodes[j].Path
	})

	return nodes, nil
}

// GetPage returns a page's content and metadata.
func (p *Provider) GetPage(_ context.Context, docName string, pagePath string) (domain.Page, error) {
	docPath := filepath.Join(p.root, docName)

	// Resolve page file path
	pageFile := filepath.Join(docPath, pagePath+".md")
	if _, err := os.Stat(pageFile); err != nil {
		// Try index.md if path is empty or ends with /
		altPath := filepath.Join(docPath, pagePath, "index.md")
		if _, err2 := os.Stat(altPath); err2 == nil {
			pageFile = altPath
		} else {
			return domain.Page{}, fmt.Errorf("%w: %s/%s", domain.ErrPageNotFound, docName, pagePath)
		}
	}

	data, err := os.ReadFile(pageFile)
	if err != nil {
		return domain.Page{}, fmt.Errorf("reading page file: %w", err)
	}

	fm, body, err := parseFrontmatter(data)
	if err != nil {
		return domain.Page{}, err
	}

	// Resolve access rules: frontmatter > access.yaml (walk up) > public default
	access, passwordHash := resolveAccess(docPath, pageFile, fm)

	page := domain.Page{
		Doc:          docName,
		Path:         pagePath,
		Title:        getString(fm, "title", filepath.Base(pagePath)),
		Description:  getString(fm, "description", ""),
		Order:        getInt(fm, "order", 0),
		Content:      body,
		Access:       access,
		PasswordHash: passwordHash,
	}

	return page, nil
}
