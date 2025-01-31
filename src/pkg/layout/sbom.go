// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: 2021-Present The Zarf Authors

// Package layout contains functions for interacting with Zarf's package layout on disk.
package layout

import (
	"fmt"
	"io/fs"
	"os"
	"path/filepath"

	"github.com/defenseunicorns/zarf/src/pkg/utils"
	"github.com/mholt/archiver/v3"
)

// ComponentSBOM contains paths for a component's SBOM.
type ComponentSBOM struct {
	Files     []string
	Component *ComponentPaths
}

// SBOMs contains paths for SBOMs.
type SBOMs struct {
	Path string
}

// Unarchive unarchives the package's SBOMs.
func (s *SBOMs) Unarchive() (err error) {
	if s.Path == "" || utils.InvalidPath(s.Path) {
		return &fs.PathError{
			Op:   "stat",
			Path: s.Path,
			Err:  fs.ErrNotExist,
		}
	}
	if utils.IsDir(s.Path) {
		return nil
	}
	tb := s.Path
	dir := filepath.Join(filepath.Dir(tb), SBOMDir)
	if err := archiver.Unarchive(tb, dir); err != nil {
		return err
	}
	s.Path = dir
	return os.Remove(tb)
}

// Archive archives the package's SBOMs.
func (s *SBOMs) Archive() (err error) {
	if s.Path == "" || utils.InvalidPath(s.Path) {
		return &fs.PathError{
			Op:   "stat",
			Path: s.Path,
			Err:  fs.ErrNotExist,
		}
	}
	if !utils.IsDir(s.Path) {
		return nil
	}
	dir := s.Path
	tb := filepath.Join(filepath.Dir(dir), SBOMTar)

	if err := utils.CreateReproducibleTarballFromDir(dir, "", tb); err != nil {
		return err
	}
	s.Path = tb
	return os.RemoveAll(dir)
}

// StageSBOMViewFiles copies SBOM viewer HTML files to the Zarf SBOM directory.
func (s *SBOMs) StageSBOMViewFiles() (warnings []string, err error) {
	if s.IsTarball() {
		return nil, fmt.Errorf("unable to process the SBOM files for this package: %s is a tarball", s.Path)
	}

	// If SBOMs were loaded, temporarily place them in the deploy directory
	if !utils.InvalidPath(s.Path) {
		if _, err := filepath.Glob(filepath.Join(s.Path, "sbom-viewer-*")); err != nil {
			return nil, err
		}

		if _, err := s.OutputSBOMFiles(SBOMDir, ""); err != nil {
			// Don't stop the deployment, let the user decide if they want to continue the deployment
			warning := fmt.Sprintf("Unable to process the SBOM files for this package: %s", err.Error())
			warnings = append(warnings, warning)
		}
	}

	return warnings, nil
}

// OutputSBOMFiles outputs SBOM files into outputDir.
func (s *SBOMs) OutputSBOMFiles(outputDir, packageName string) (string, error) {
	packagePath := filepath.Join(outputDir, packageName)

	if err := os.RemoveAll(packagePath); err != nil {
		return "", err
	}

	if err := utils.CreateDirectory(packagePath, 0700); err != nil {
		return "", err
	}

	return packagePath, utils.CreatePathAndCopy(s.Path, packagePath)
}

// IsTarball returns true if the SBOMs are a tarball.
func (s SBOMs) IsTarball() bool {
	return !utils.IsDir(s.Path) && filepath.Ext(s.Path) == ".tar"
}
