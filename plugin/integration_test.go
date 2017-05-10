// +build integration

package plugin

import (
	"fmt"
	"os/exec"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/themotion/ladder/config"
)

func TestLoadSingle(t *testing.T) {
	assert := assert.New(t)
	// Compile plugins on tmp
	plgPath := "/tmp/plugin1.so"
	cmdArgs := fmt.Sprintf("build -buildmode=plugin -o %s testdata/plugin1.go", plgPath)
	cmd := exec.Command("go", strings.Split(cmdArgs, " ")...)
	err := cmd.Run()

	// Check plugin can be loaded
	if assert.NoError(err, "Plugin compilation shouldn't return an error") {
		pLoader, err := NewBaseLoader()
		if assert.NoError(err) {
			plg, err := pLoader.Load(plgPath)
			assert.NoError(err, "Plugin load shouldn't error")
			assert.NotNil(plg)
			sum, err := plg.Lookup("Sum")
			assert.NoError(err, "Sum function should be implemented by the plugin")
			res := sum.(func(int, int) int)(5, 4)
			assert.EqualValues(9, res)
		}
	}
}

func TestLoadFromCfg(t *testing.T) {
	assert := assert.New(t)
	// Compile plugins on tmp
	plg1Path := "/tmp/plugin1.so"
	plg2Path := "/tmp/plugin2.so"
	cmdArgs := "build -buildmode=plugin -o %s %s"
	cmd1Args := fmt.Sprintf(cmdArgs, plg1Path, "testdata/plugin1.go")
	cmd2Args := fmt.Sprintf(cmdArgs, plg2Path, "testdata/plugin2.go")

	cmd1 := exec.Command("go", strings.Split(cmd1Args, " ")...)
	cmd2 := exec.Command("go", strings.Split(cmd2Args, " ")...)

	if assert.NoError(cmd1.Run(), "Plugin 1 compilation shouldn't fail") &&
		assert.NoError(cmd2.Run(), "Plugin 2 compilation shouldn't fail") {
		// Check plugin can be loaded
		pLoader, err := NewBaseLoader()
		if assert.NoError(err) {
			// Create our loader configuration
			cfg := config.Config{
				Global: config.Global{
					Plugins: []string{
						plg1Path,
						plg2Path,
					},
				},
			}

			// Load configuration and check if all the plugins have been loaded
			err := pLoader.LoadFromConfig(&cfg)
			assert.NoError(err, "Plugins load shouldn't error")
			if assert.Len(pLoader.Plugins, 2, "There should be 2 plugins loaded") {
				// Check first plugin
				sum, err := pLoader.Plugins[plg1Path].Lookup("Sum")
				assert.NoError(err, "Sum function should be implemented by the plugin")
				res := sum.(func(int, int) int)(5, 4)
				assert.EqualValues(9, res)

				// Check second plugin
				sub, err := pLoader.Plugins[plg2Path].Lookup("Sub")
				assert.NoError(err, "Sub function should be implemented by the plugin")
				res = sub.(func(int, int) int)(5, 4)
				assert.EqualValues(1, res)
			}

		}
	}
}
