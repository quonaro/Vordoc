package content

import (
	"os"
	"path/filepath"
	"sort"
	"strings"

	"vordoc/internal/domain"
)

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
			show := true
			idx := filepath.Join(fullPath, "index.md")
			if data, err := os.ReadFile(idx); err == nil { // #nosec G304 — path is built internally
				hasIndex = true
				fm, _, _ := parseFrontmatter(data)
				if pc, found := loadPageConfigEntry(docPath, idx); found {
					fm = mergePageConfig(fm, pc)
				}
				if t := getString(fm, "title", ""); t != "" {
					node.Title = t
				}
				node.Icon = getString(fm, "icon", "")
				node.Order = getInt(fm, "order", 0)
				info := resolveAccessInfo(docPath, idx, fm)
				node.Access = info.Access
				node.AccessScope = info.Scope
				node.LockColor = hashToColor(info.PasswordHash)
				show = getBool(fm, "show", true)
			}
			if node.Icon == "" {
				if cfg, err := loadDocConfig(filepath.Join(fullPath, "config.yaml")); err == nil {
					node.Icon = cfg.Icon
				}
			}
			if !show {
				continue
			}
			if len(children) == 0 && !hasIndex {
				continue
			}
			node.HasIndex = hasIndex
			node.Show = show
			nodes = append(nodes, node)
		} else if filepath.Ext(name) == ".md" {
			if name == "index.md" && dir == docPath {
				continue
			}
			if name == "index.md" {
				continue
			}

			data, err := os.ReadFile(fullPath) // #nosec G304 — path is built internally
			if err != nil {
				continue
			}
			fm, _, _ := parseFrontmatter(data)
			if pc, found := loadPageConfigEntry(docPath, fullPath); found {
				fm = mergePageConfig(fm, pc)
			}
			title := getString(fm, "title", strings.TrimSuffix(name, ".md"))
			icon := getString(fm, "icon", "")
			order := getInt(fm, "order", 0)
			info := resolveAccessInfo(docPath, fullPath, fm)
			show := getBool(fm, "show", true)
			if !show {
				continue
			}
			nodes = append(nodes, domain.PageNode{
				Path:        strings.TrimSuffix(rel, ".md"),
				Title:       title,
				Icon:        icon,
				Order:       order,
				Access:      info.Access,
				AccessScope: info.Scope,
				LockColor:   hashToColor(info.PasswordHash),
				Show:        show,
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
