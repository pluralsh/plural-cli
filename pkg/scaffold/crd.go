package scaffold

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"

	"github.com/pluralsh/plural/pkg/api"
	"github.com/pluralsh/plural/pkg/utils"
	"github.com/pluralsh/plural/pkg/utils/pathing"
	"github.com/pluralsh/plural/pkg/wkspace"
)

func (s *Scaffold) buildCrds(wk *wkspace.Workspace) error {
	utils.Highlight("syncing crds")
	if err := os.RemoveAll(s.Root); err != nil {
		return err
	}

	if err := os.MkdirAll(s.Root, os.ModePerm); err != nil {
		return err
	}

	for _, chartInst := range wk.Charts {
		for _, crd := range chartInst.Version.Crds {
			utils.Highlight(".")
			if err := writeCrd(s.Root, &crd); err != nil {
				fmt.Print("\n")
				return err
			}
		}
	}

	utils.Success("\u2713\n")
	return nil
}

func writeCrd(path string, crd *api.Crd) error {
	resp, err := http.Get(crd.Blob)
	if err != nil {
		return err
	}

	contents, err := io.ReadAll(resp.Body)
	if err != nil {
		return err
	}

	return os.WriteFile(pathing.SanitizeFilepath(filepath.Join(path, crd.Name)), contents, 0644)
}
