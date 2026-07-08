package content

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
	"unicode"
	"unicode/utf8"

	"vordoc/internal/domain"
)

// SearchPages searches for query across all pages within the given documentation.
func (p *Provider) SearchPages(_ context.Context, docName string, query string) ([]domain.SearchResult, error) {
	query = strings.TrimSpace(query)
	if query == "" {
		return nil, nil
	}

	docPath := filepath.Join(p.root, docName)
	info, err := os.Stat(docPath)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, fmt.Errorf("%w: %s", domain.ErrDocNotFound, docName)
		}
		return nil, fmt.Errorf("stat doc: %w", err)
	}
	if !info.IsDir() {
		return nil, fmt.Errorf("%w: %s is not a directory", domain.ErrDocNotFound, docName)
	}

	terms := searchTerms(query)
	if len(terms) == 0 {
		return nil, nil
	}

	var results []domain.SearchResult

	walkErr := filepath.Walk(docPath, func(fullPath string, fi os.FileInfo, err error) error {
		if err != nil {
			return nil
		}
		if fi.IsDir() || filepath.Ext(fullPath) != ".md" {
			return nil
		}

		rel, _ := filepath.Rel(docPath, fullPath)
		rel = filepath.ToSlash(rel)

		data, err := os.ReadFile(fullPath)
		if err != nil {
			return nil
		}

		fm, body, _ := parseFrontmatter(data)
		title := getString(fm, "title", "")
		accessInfo := resolveAccessInfo(docPath, fullPath, fm)

		// Determine page path used in URLs.
		pagePath := strings.TrimSuffix(rel, ".md")
		if pagePath == "index" {
			pagePath = ""
		}
		// If title is missing, derive from the path segment.
		if title == "" {
			title = filepath.Base(pagePath)
		}

		cleanText := stripMarkdown(body)
		score := scoreTerms(terms, title, pagePath, cleanText)
		if score == 0 {
			return nil
		}

		results = append(results, domain.SearchResult{
			Title:       title,
			Path:        pagePath,
			Snippet:     snippet(cleanText, terms),
			Access:      accessInfo.Access,
			AccessScope: accessInfo.Scope,
			Score:       score,
		})
		return nil
	})
	if walkErr != nil {
		return nil, fmt.Errorf("walking doc pages: %w", walkErr)
	}

	// Higher score first, then title alphabetical for stable ordering.
	sort.Slice(results, func(i, j int) bool {
		if results[i].Score != results[j].Score {
			return results[i].Score > results[j].Score
		}
		return results[i].Title < results[j].Title
	})

	return results, nil
}

// SearchAllDocs searches for query across all available documentations.
func (p *Provider) SearchAllDocs(ctx context.Context, query string) ([]domain.GlobalSearchResult, error) {
	query = strings.TrimSpace(query)
	if query == "" {
		return nil, nil
	}

	docNames, err := p.ListDocs(ctx)
	if err != nil {
		return nil, fmt.Errorf("listing docs for global search: %w", err)
	}

	const maxPagesPerDoc = 10

	var all []domain.GlobalSearchResult
	for _, name := range docNames {
		summary, err := p.GetDocSummary(ctx, name)
		if err != nil {
			continue
		}

		group := domain.GlobalSearchResult{
			Name:        summary.Name,
			Title:       summary.Title,
			Description: summary.Description,
			Access:      summary.Access,
		}

		if summary.Access != "password" {
			pages, err := p.SearchPages(ctx, name, query)
			if err != nil {
				continue
			}
			if len(pages) == 0 {
				continue
			}
			if len(pages) > maxPagesPerDoc {
				pages = pages[:maxPagesPerDoc]
			}
			for _, page := range pages {
				group.Pages = append(group.Pages, domain.GlobalSearchPageResult{
					Title:   page.Title,
					Path:    page.Path,
					Snippet: page.Snippet,
				})
			}
		}

		all = append(all, group)
	}

	sort.Slice(all, func(i, j int) bool {
		return strings.ToLower(all[i].Title) < strings.ToLower(all[j].Title)
	})

	return all, nil
}

// searchTerms splits a query into normalized lowercase tokens, removing punctuation.
func searchTerms(query string) []string {
	var terms []string
	for _, raw := range strings.Fields(query) {
		raw = strings.ToLower(raw)
		raw = strings.TrimFunc(raw, func(r rune) bool {
			return unicode.IsPunct(r) || unicode.IsSymbol(r)
		})
		if raw != "" && len(raw) >= 2 {
			terms = append(terms, raw)
		}
	}
	return terms
}

// scoreTerms returns a non-zero score if all terms match at least one field.
// Title matches contribute the most, then path, then content.
func scoreTerms(terms []string, title, path, text string) int {
	lowerTitle := strings.ToLower(title)
	lowerPath := strings.ToLower(path)
	lowerText := strings.ToLower(text)

	score := 0
	allMatch := true
	for _, term := range terms {
		matched := false
		if strings.Contains(lowerTitle, term) {
			score += 40
			matched = true
		}
		if strings.Contains(lowerPath, term) {
			score += 20
			matched = true
		}
		if strings.Contains(lowerText, term) {
			score += 10
			matched = true
		}
		if !matched {
			allMatch = false
			break
		}
	}
	if !allMatch {
		return 0
	}
	return score
}

var (
	// linkRegex removes markdown links keeping link text: [text](url) -> text
	linkRegex = regexp.MustCompile(`\[([^\]]*)\]\([^)]*\)`)
	// imageRegex removes markdown images: ![alt](url) -> ""
	imageRegex = regexp.MustCompile(`!\[[^\]]*\]\([^)]*\)`)
	// htmlTagRegex removes HTML tags.
	htmlTagRegex = regexp.MustCompile(`<[^>]+>`)
	// codeFenceRegex matches fenced code blocks and captures their inner content.
	codeFenceRegex = regexp.MustCompile("(?s)```[ \t]*[^\n]*\n?(.*?)```")
	// inlineCodeRegex removes inline code spans keeping the content.
	inlineCodeRegex = regexp.MustCompile("`([^`]*)`")
	// headingRegex removes leading hash marks from headings.
	headingRegex = regexp.MustCompile(`(?m)^#{1,6}\s*`)
	// emphasisRegex removes emphasis markers (* and _).
	emphasisRegex = regexp.MustCompile(`(\*+|_+)`)
	// bulletRegex removes bullet list markers.
	bulletRegex = regexp.MustCompile(`(?m)^\s*[-*+]\s*`)
	// numberedRegex removes numbered list markers.
	numberedRegex = regexp.MustCompile(`(?m)^\s*\d+\.\s*`)
	// blockquoteRegex removes blockquote markers.
	blockquoteRegex = regexp.MustCompile(`(?m)^\s*>\s*`)
	// whitespaceRegex normalizes whitespace.
	whitespaceRegex = regexp.MustCompile(`\s+`)
)

// stripMarkdown returns a plain text approximation of a markdown body.
func stripMarkdown(markdown string) string {
	// Replace fenced code blocks with placeholders so their content is not
	// mangled by inline-code or emphasis rules.
	placeholders := []string{}
	out := codeFenceRegex.ReplaceAllStringFunc(markdown, func(match string) string {
		parts := codeFenceRegex.FindStringSubmatch(match)
		placeholder := fmt.Sprintf("VORDOC-CODEBLOCK-%d", len(placeholders))
		if len(parts) > 1 {
			placeholders = append(placeholders, parts[1])
		} else {
			placeholders = append(placeholders, "")
		}
		return "\n" + placeholder + "\n"
	})
	out = imageRegex.ReplaceAllString(out, "")
	out = linkRegex.ReplaceAllString(out, "$1")
	out = inlineCodeRegex.ReplaceAllString(out, "$1")
	out = htmlTagRegex.ReplaceAllString(out, "")
	out = headingRegex.ReplaceAllString(out, "")
	out = blockquoteRegex.ReplaceAllString(out, "")
	out = bulletRegex.ReplaceAllString(out, "")
	out = numberedRegex.ReplaceAllString(out, "")
	out = emphasisRegex.ReplaceAllString(out, "")
	for i, block := range placeholders {
		placeholder := fmt.Sprintf("VORDOC-CODEBLOCK-%d", i)
		out = strings.Replace(out, placeholder, "\n"+block+"\n", 1)
	}
	out = whitespaceRegex.ReplaceAllString(out, " ")
	out = strings.TrimSpace(out)
	return out
}

// snippet extracts a short excerpt around the first term match.
// It slices by rune boundaries so multi-byte characters are not broken.
func snippet(text string, terms []string) string {
	lowerText := strings.ToLower(text)
	runes := []rune(text)
	idx := -1
	for _, term := range terms {
		i := strings.Index(lowerText, term)
		if i != -1 {
			runeIdx := utf8.RuneCountInString(lowerText[:i])
			if idx == -1 || runeIdx < idx {
				idx = runeIdx
			}
		}
	}
	if idx == -1 {
		idx = 0
	}

	start := idx - 60
	if start < 0 {
		start = 0
	}
	end := idx + 140
	if end > len(runes) {
		end = len(runes)
	}

	excerpt := string(runes[start:end])
	if start > 0 {
		excerpt = "..." + excerpt
	}
	if end < len(runes) {
		excerpt = excerpt + "..."
	}
	return whitespaceRegex.ReplaceAllString(excerpt, " ")
}
