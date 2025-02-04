package update

import (
	"fmt"
	"os"
	"path/filepath"

	"chainguard.dev/melange/pkg/renovate"
	"chainguard.dev/melange/pkg/renovate/bump"

	"chainguard.dev/melange/pkg/build"
	"github.com/pkg/errors"
	"gopkg.in/yaml.v2"
)

type MelageConfig struct {
	Package  build.Package    `yaml:"package"`
	Pipeline []build.Pipeline `yaml:"pipeline,omitempty"`
}

func (o Options) readAllPackagesFromRepo(tempDir string) (map[string]MelageConfig, error) {
	var fileList []string
	packageConfigs := make(map[string]MelageConfig)
	err := filepath.Walk(tempDir, func(path string, fi os.FileInfo, err error) error {
		// skip if the path is not the root folder of the melange config repo
		if fi.IsDir() && path != tempDir {
			return filepath.SkipDir
		}
		if filepath.Ext(path) == ".yaml" {
			fileList = append(fileList, path)
		}
		return nil
	})

	if err != nil {
		return packageConfigs, errors.Wrapf(err, "failed walking files in cloned directory %s", tempDir)
	}

	fmt.Printf("found %[1]d packages\n", len(fileList))

	for _, fi := range fileList {
		packageConfig, err := o.readPackageConfig(fi)
		if err != nil {
			return packageConfigs, errors.Wrapf(err, "failed to read package config %s", fi)
		}

		packageConfigs[packageConfig.Package.Name] = packageConfig
	}
	return packageConfigs, nil
}

// read a single melange config using the package name to match the filename
func (o Options) readPackageConfig(filename string) (MelageConfig, error) {
	packageConfig := MelageConfig{}

	data, err := os.ReadFile(filename)
	if err != nil {
		return packageConfig, errors.Wrapf(err, "failed to read package config %s", filename)
	}

	err = yaml.Unmarshal(data, &packageConfig)
	if err != nil {
		return packageConfig, errors.Wrapf(err, "failed to unmarshal package data from filename %s", filename)
	}

	return packageConfig, nil
}

func (o Options) bump(configFile, version string) error {
	ctx, err := renovate.New(renovate.WithConfig(configFile))
	if err != nil {
		return err
	}

	bumpRenovator := bump.New(
		bump.WithTargetVersion(version),
	)

	if err := ctx.Renovate(bumpRenovator); err != nil {
		return err
	}
	return nil
}
