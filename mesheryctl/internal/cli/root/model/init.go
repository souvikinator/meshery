package model

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/layer5io/meshery/mesheryctl/pkg/utils"
	"github.com/layer5io/meshkit/models/meshmodel/entity"
	"github.com/meshery/schemas/models/v1beta1"
	_model "github.com/meshery/schemas/models/v1beta1/model"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
)

var (
	outputFormat string
	version      string
	outputPath   string
)

var (
	ErrModelInitCode = "replace_me"
)

var initCmd = &cobra.Command{
	Use:   "init [model name]",
	Short: "Initialize a new Meshery model",
	Long: `Initialize a new Meshery model with proper scaffolding.
Creates a version-aware directory structure with sample components and relationships.`,
	Example: `
mesheryctl model init my-model
mesheryctl model init my-model --output-format yaml
mesheryctl model init my-model --version 1.0.0
`,
	PreRunE: func(cmd *cobra.Command, args []string) error {
		if len(args) != 1 {
			return utils.ErrModelInit(nil, "Model name is required")
		}

		if !isValidModelName(args[0]) {
			return utils.ErrModelInit(nil, "Invalid model name. Must be lowercase with hyphens only")
		}

		if !isValidFormat(outputFormat) {
			return utils.ErrModelInit(nil, "Invalid output format. Must be json, yaml, or csv")
		}

		if !isValidVersion(version) {
			return utils.ErrModelInit(nil, "Invalid version format. Must be semver (e.g., 1.0.0)")
		}

		return nil
	},
	RunE: func(cmd *cobra.Command, args []string) error {
		modelName := args[0]

		err := createModelScaffolding(modelName, version, outputFormat, outputPath)
		if err != nil {
			return utils.ErrModelInit(err, "Failed to create model scaffolding")
		}

		// Display success and next steps
		displayNextSteps(modelName, version)
		return nil
	},
}

func init() {
	initCmd.Flags().SetNormalizeFunc(func(f *pflag.FlagSet, name string) pflag.NormalizedName {
		return pflag.NormalizedName(strings.ToLower(name))
	})

	initCmd.Flags().StringVarP(&outputFormat, "output-format", "o", "json", "Output format (json|yaml|csv)")
	initCmd.Flags().StringVarP(&version, "version", "v", "1.0.0", "Model version")
	initCmd.Flags().StringVarP(&outputPath, "path", "p", "", "Target directory (default: current directory)")
}

func isValidModelName(name string) bool {
	// Only lowercase letters, numbers and hyphens
	return regexp.MustCompile(`^[a-z0-9][a-z0-9-]*$`).MatchString(name)
}

func isValidFormat(format string) bool {
	validFormats := map[string]bool{
		"json": true,
		"yaml": true,
		"csv":  true,
	}
	return validFormats[strings.ToLower(format)]
}

func isValidVersion(v string) bool {
	// Basic semver validation
	return regexp.MustCompile(`^v?\d+\.\d+\.\d+$`).MatchString(v)
}

func createModelScaffolding(name, version, format, path string) error {
	// If no path specified, use current directory
	if path == "" {
		var err error
		path, err = os.Getwd()
		if err != nil {
			return utils.ErrModelInit(err, "Failed to get current directory")
		}
	}

	// Create the model definition
	modelDef, err := createModelDefinition(name, version)
	if err != nil {
		return utils.ErrModelInit(err, "Failed to create model definition")
	}

	// Create versioned directory structure
	modelDir, compDir, relDir, err := createVersionedDirectories(path, name, version)
	if err != nil {
		return utils.ErrModelInit(err, "Failed to create directory structure")
	}

	// Write model definition to file
	err = writeModelDefinition(modelDef, modelDir, format)
	if err != nil {
		return utils.ErrModelInit(err, "Failed to write model definition")
	}

	// Create sample components and relationships (no-op for now)
	err = createSampleContent(compDir, relDir, format)
	if err != nil {
		return utils.ErrModelInit(err, "Failed to create sample content")
	}

	return nil
}

func createModelDefinition(name, version string) (*_model.ModelDefinition, error) {
	// Create a basic model definition with sensible defaults
	model := _model.ModelDefinition{
		Name:          name,
		Version:       version,
		DisplayName:   strings.Title(strings.ReplaceAll(name, "-", " ")), // Convert my-model to "My Model"
		Status:        _model.ModelDefinitionStatus(entity.Enabled),
		SchemaVersion: v1beta1.ModelSchemaVersion,
		Model: _model.Model{
			Version: version,
		},
		// Minimal metadata
		Metadata: &_model.ModelDefinition_Metadata{
			AdditionalProperties: make(map[string]interface{}),
		},
	}

	return &model, nil
}

func createVersionedDirectories(basePath, modelName, version string) (modelDir, compDir, relDir string, err error) {
	// Ensure version has v prefix
	if !strings.HasPrefix(version, "v") {
		version = "v" + version
	}

	// Create base model directory
	modelDir = filepath.Join(basePath, modelName, version)
	if err = os.MkdirAll(modelDir, 0755); err != nil {
		return "", "", "", err
	}

	// Create components directory
	compDir = filepath.Join(modelDir, "components")
	if err = os.MkdirAll(compDir, 0755); err != nil {
		return "", "", "", err
	}

	// Create relationships directory
	relDir = filepath.Join(modelDir, "relationships")
	if err = os.MkdirAll(relDir, 0755); err != nil {
		return "", "", "", err
	}

	return modelDir, compDir, relDir, nil
}

func writeModelDefinition(model *_model.ModelDefinition, modelDir, format string) error {
	var ext string
	switch strings.ToLower(format) {
	case "json":
		ext = "json"
	case "yaml":
		ext = "yaml"
	case "csv":
		ext = "csv"
	default:
		return fmt.Errorf("unsupported format: %s", format)
	}

	filePath := filepath.Join(modelDir, fmt.Sprintf("model.%s", ext))

	// Use existing WriteModelDefinition method from the model package
	return model.WriteModelDefinition(filePath, format)
}

func displayNextSteps(name, version string) {
	utils.Log.Info("Successfully created model scaffolding")
	utils.Log.Info("\nNext steps:")
	utils.Log.Info(fmt.Sprintf("1. cd %s/v%s", name, version))
	utils.Log.Info("2. Edit model files")
	utils.Log.Info("3. Add components and relationships")
	utils.Log.Info("\nImport model:")
	utils.Log.Info(fmt.Sprintf("  mesheryctl model import ./%s/v%s", name, version))
	utils.Log.Info("\nBuild model:")
	utils.Log.Info(fmt.Sprintf("  mesheryctl model build ./%s/v%s", name, version))
}
