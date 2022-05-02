package manifest

import (
	"path/filepath"
)

type Links struct {
	Terraform map[string]string
	Helm      map[string]string
}

func (man *Manifest) AddLink(tool, name, path string) {
	links := man.Links
	if links == nil {
		links = &Links{
			Terraform: map[string]string{},
			Helm:      map[string]string{},
		}
	}

	absPath, _ := filepath.Abs(path)

	if tool == "terraform" {
		links.Terraform[name] = absPath
	}

	if tool == "helm" {
		links.Helm[name] = absPath
	}

	man.Links = links
}

func (man *Manifest) Unlink(tool, name string) {
	links := man.Links
	if links == nil {
		return
	}

	if tool == "terraform" {
		delete(links.Terraform, name)
	} else if tool == "helm" {
		delete(links.Helm, name)
	}

	man.Links = links
}

func (man *Manifest) UnlinkAll() {
	man.Links = nil
}
