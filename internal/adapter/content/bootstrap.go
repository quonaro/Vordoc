package content

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
)

const defaultRootConfigYAML = `# Vordoc root site configuration.
# This file controls the top-level documentation site and provides defaults
# that can be overridden by each documentation's config.yaml.
#
# root:             root landing page settings
#   enable:         show a root page listing all documentations (true/false)
#   title:          title shown on the root page
# favicon:          path to the favicon, relative to the frontend public directory
# header:           site-wide header settings
#   enable:         show the header (true/false)
#   selector:       show the theme selector in the header (true/false)
#   title:          site title used in the header
#   logo:           logo settings
#     path:         path to the logo file, relative to the content root
#     size:         logo height in pixels
#   font:           header title font settings
#     name:         standard font family or path to a .ttf/.otf font file
#     size:         title font size in pixels
# theme:            site-wide theme selector settings
#   default:        default theme: system, light, or dark
#   accent-color:   accent color as a hex value (e.g. "#3b82f6")
root:
  enable: true
  title: "Vordoc"

favicon: "favicon.ico"

header:
  enable: true
  selector: true
  title: "Vordoc"
  logo:
    path: "logotype.svg"
    size: 40
  font:
    name: "FabergeDigital.otf"
    size: 24

theme:
  default: "system"
  accent-color: "#3b82f6"
`

const defaultLogoSVG = `<svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 32 32" fill="none" role="img">
  <rect width="32" height="32" rx="6" fill="#3b82f6"/>
  <path d="M9 11 L16 21 L23 11" stroke="white" stroke-width="3" stroke-linecap="round" stroke-linejoin="round"/>
</svg>
`

// EnsureDefaults creates the content root directory, root config.yaml, and the
// default logo file when they are missing. Existing files are never overwritten.
func (p *Provider) EnsureDefaults(ctx context.Context) error {
	if err := os.MkdirAll(p.root, 0o755); err != nil {
		return fmt.Errorf("creating content root: %w", err)
	}
	if err := p.ensureRootConfig(); err != nil {
		return fmt.Errorf("ensuring root config: %w", err)
	}
	if err := p.ensureDefaultLogo(); err != nil {
		return fmt.Errorf("ensuring default logo: %w", err)
	}
	return nil
}

func (p *Provider) ensureRootConfig() error {
	path := filepath.Join(p.root, "config.yaml")
	if _, err := os.Stat(path); err == nil {
		return nil
	} else if !os.IsNotExist(err) {
		return err
	}
	return os.WriteFile(path, []byte(defaultRootConfigYAML), 0o644)
}

func (p *Provider) ensureDefaultLogo() error {
	path := filepath.Join(p.root, defaultLogoFile)
	if _, err := os.Stat(path); err == nil {
		return nil
	} else if !os.IsNotExist(err) {
		return err
	}
	return os.WriteFile(path, []byte(defaultLogoSVG), 0o644)
}
