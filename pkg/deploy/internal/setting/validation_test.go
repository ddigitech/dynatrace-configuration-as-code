/*
 * @license
 * Copyright 2024 Dynatrace LLC
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 * http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package setting

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/dynatrace/dynatrace-configuration-as-code/v2/pkg/config"
	"github.com/dynatrace/dynatrace-configuration-as-code/v2/pkg/config/coordinate"
	"github.com/dynatrace/dynatrace-configuration-as-code/v2/pkg/config/parameter"
	refParam "github.com/dynatrace/dynatrace-configuration-as-code/v2/pkg/config/parameter/reference"
	valueParam "github.com/dynatrace/dynatrace-configuration-as-code/v2/pkg/config/parameter/value"
	project "github.com/dynatrace/dynatrace-configuration-as-code/v2/pkg/project/v2"
)

func TestInsertAfterSameScopeValidator(t *testing.T) {

	validator := InsertAfterSameScopeValidator{}

	tests := []struct {
		name                string
		sourceConfig        config.Config
		otherProjectConfigs []config.Config
		expectError         error // if nil -> no error expected
	}{
		{
			name: "Valid single config",
			sourceConfig: config.Config{
				Coordinate:  coordinate.Coordinate{Project: "project-id", Type: "type-a", ConfigId: "config-x"},
				Environment: "env",
				Type:        &config.SettingsType{SchemaId: "type-a"},
			},
			otherProjectConfigs: []config.Config{},
		},
		{
			name: "Valid reference to other config",
			sourceConfig: config.Config{
				Coordinate:  coordinate.Coordinate{Project: "project-id", Type: "type-a", ConfigId: "config-x"},
				Environment: "env",
				Type:        &config.SettingsType{SchemaId: "type-a"},
				Parameters: map[string]parameter.Parameter{
					config.InsertAfterParameter: refParam.New("my-project", "type-a", "config-y", "id"),
					config.ScopeParameter:       valueParam.New("environment"),
				},
			},
			otherProjectConfigs: []config.Config{
				{
					Coordinate:  coordinate.Coordinate{Project: "project-id", Type: "type-a", ConfigId: "config-y"},
					Environment: "env",
					Type:        &config.SettingsType{SchemaId: "type-a"},
					Parameters: map[string]parameter.Parameter{
						config.ScopeParameter: valueParam.New("environment"),
					},
				},
			},
		},
		{
			name: "Referenced config has different scope",
			sourceConfig: config.Config{
				Coordinate:  coordinate.Coordinate{Project: "project-id", Type: "type-a", ConfigId: "config-x"},
				Environment: "env",
				Type:        &config.SettingsType{SchemaId: "type-a"},
				Parameters: map[string]parameter.Parameter{
					config.InsertAfterParameter: refParam.New("my-project", "type-a", "config-y", "id"),
					config.ScopeParameter:       valueParam.New("environment"),
				},
			},
			otherProjectConfigs: []config.Config{
				{
					Coordinate:  coordinate.Coordinate{Project: "project-id", Type: "type-a", ConfigId: "config-y"},
					Environment: "env",
					Type:        &config.SettingsType{SchemaId: "type-a"},
					Parameters: map[string]parameter.Parameter{
						config.ScopeParameter: valueParam.New("entity"),
					},
				},
			},
			expectError: errDiffScope,
		},
		{
			name: "Referenced config does not exist at all",
			sourceConfig: config.Config{
				Coordinate:  coordinate.Coordinate{Project: "project-id", Type: "type-a", ConfigId: "config-x"},
				Environment: "env",
				Type:        &config.SettingsType{SchemaId: "type-a"},
				Parameters: map[string]parameter.Parameter{
					config.InsertAfterParameter: refParam.New("my-project", "type-a", "config-y", "id"),
					config.ScopeParameter:       valueParam.New("environment"),
				},
			},
			otherProjectConfigs: []config.Config{},
			expectError:         errReferencedNotFound,
		},
		{
			name: "Referenced config exists only in other env",
			sourceConfig: config.Config{
				Coordinate:  coordinate.Coordinate{Project: "project-id", Type: "type-a", ConfigId: "config-x"},
				Environment: "env",
				Type:        &config.SettingsType{SchemaId: "type-a"},
				Parameters: map[string]parameter.Parameter{
					config.InsertAfterParameter: refParam.New("my-project", "type-a", "config-y", "id"),
					config.ScopeParameter:       valueParam.New("environment"),
				},
			},
			otherProjectConfigs: []config.Config{
				{
					Coordinate:  coordinate.Coordinate{Project: "project-id", Type: "type-a", ConfigId: "config-y"},
					Environment: "other-env", // instead of 'env'
					Type:        &config.SettingsType{SchemaId: "type-a"},
					Parameters: map[string]parameter.Parameter{
						config.ScopeParameter: valueParam.New("entity"),
					},
				},
			},
			expectError: errReferencedNotFound,
		},
		{
			name: "Referenced config exists but schema is different",
			sourceConfig: config.Config{
				Coordinate:  coordinate.Coordinate{Project: "project-id", Type: "type-a", ConfigId: "config-x"},
				Environment: "env",
				Type:        &config.SettingsType{SchemaId: "type-a"},
				Parameters: map[string]parameter.Parameter{
					config.InsertAfterParameter: refParam.New("my-project", "type-b", "config-y", "id"),
					config.ScopeParameter:       valueParam.New("environment"),
				},
			},
			otherProjectConfigs: []config.Config{},
			expectError:         errDiffSchema,
		},
		{
			name: "Referenced config does not exist, but config is skipped so no error",
			sourceConfig: config.Config{
				Coordinate:  coordinate.Coordinate{Project: "project-id", Type: "type-a", ConfigId: "config-x"},
				Environment: "env",
				Type:        &config.SettingsType{SchemaId: "type-a"},
				Skip:        true,
				Parameters: map[string]parameter.Parameter{
					config.InsertAfterParameter: refParam.New("my-project", "type-a", "config-y", "id"),
					config.ScopeParameter:       valueParam.New("environment"),
				},
			},
			otherProjectConfigs: []config.Config{},
			expectError:         nil,
		},
		{
			name: "InsertAfter is not a reference, so no check can be performed",
			sourceConfig: config.Config{
				Coordinate:  coordinate.Coordinate{Project: "project-id", Type: "type-a", ConfigId: "config-x"},
				Environment: "env",
				Type:        &config.SettingsType{SchemaId: "type-a"},
				Parameters: map[string]parameter.Parameter{
					config.InsertAfterParameter: valueParam.New("static-reference"),
				},
			},
			otherProjectConfigs: []config.Config{},
			expectError:         nil,
		},
		{
			name: "config is not a settings config, so no validation is performed",
			sourceConfig: config.Config{
				Type: &config.ClassicApiType{},
			},
			otherProjectConfigs: []config.Config{},
			expectError:         nil,
		},
		{
			name: "Valid reference to other config but source scope is not a value parameter, so the check can't be performed",
			sourceConfig: config.Config{
				Coordinate:  coordinate.Coordinate{Project: "project-id", Type: "type-a", ConfigId: "config-x"},
				Environment: "env",
				Type:        &config.SettingsType{SchemaId: "type-a"},
				Parameters: map[string]parameter.Parameter{
					config.InsertAfterParameter: refParam.New("my-project", "type-a", "config-y", "id"),
					config.ScopeParameter:       refParam.New("", "", "", ""),
				},
			},
			otherProjectConfigs: []config.Config{
				{
					Coordinate:  coordinate.Coordinate{Project: "project-id", Type: "type-a", ConfigId: "config-y"},
					Environment: "env",
					Type:        &config.SettingsType{SchemaId: "type-a"},
					Parameters: map[string]parameter.Parameter{
						config.ScopeParameter: valueParam.New("environment"),
					},
				},
			},
		},
		{
			name: "Valid reference to other config but target scope is not a value parameter, so the check can't be performed",
			sourceConfig: config.Config{
				Coordinate:  coordinate.Coordinate{Project: "project-id", Type: "type-a", ConfigId: "config-x"},
				Environment: "env",
				Type:        &config.SettingsType{SchemaId: "type-a"},
				Parameters: map[string]parameter.Parameter{
					config.InsertAfterParameter: refParam.New("my-project", "type-a", "config-y", "id"),
					config.ScopeParameter:       valueParam.New("environment"),
				},
			},
			otherProjectConfigs: []config.Config{
				{
					Coordinate:  coordinate.Coordinate{Project: "project-id", Type: "type-a", ConfigId: "config-y"},
					Environment: "env",
					Type:        &config.SettingsType{SchemaId: "type-a"},
					Parameters: map[string]parameter.Parameter{
						config.ScopeParameter: refParam.New("", "", "", ""),
					},
				},
			},
		},
		{
			name: "Valid reference to other config but target scope is not a simple string parameter, so the check can't be performed",
			sourceConfig: config.Config{
				Coordinate:  coordinate.Coordinate{Project: "project-id", Type: "type-a", ConfigId: "config-x"},
				Environment: "env",
				Type:        &config.SettingsType{SchemaId: "type-a"},
				Parameters: map[string]parameter.Parameter{
					config.InsertAfterParameter: refParam.New("my-project", "type-a", "config-y", "id"),
					config.ScopeParameter:       valueParam.New("environment"),
				},
			},
			otherProjectConfigs: []config.Config{
				{
					Coordinate:  coordinate.Coordinate{Project: "project-id", Type: "type-a", ConfigId: "config-y"},
					Environment: "env",
					Type:        &config.SettingsType{SchemaId: "type-a"},
					Parameters: map[string]parameter.Parameter{
						config.ScopeParameter: valueParam.New(map[string]any{}),
					},
				},
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			proj := buildProject(test.sourceConfig, test.otherProjectConfigs)

			err := validator.Validate(proj, test.sourceConfig)
			if test.expectError == nil {
				assert.NoError(t, err)
			} else {
				assert.Error(t, err)
				assert.ErrorIs(t, err, test.expectError)
			}
		})
	}
}

func buildProject(sourceConfig config.Config, configs []config.Config) project.Project {

	configs = append(configs, sourceConfig)

	projectConfigs := project.ConfigsPerTypePerEnvironments{}
	for _, c := range configs {
		if _, f := projectConfigs[c.Environment]; !f {
			projectConfigs[c.Environment] = map[project.ConfigTypeName][]config.Config{}
		}

		if _, f := projectConfigs[c.Environment][c.Coordinate.Type]; !f {
			projectConfigs[c.Environment][c.Coordinate.Type] = []config.Config{}
		}

		projectConfigs[c.Environment][c.Coordinate.Type] = append(projectConfigs[c.Environment][c.Coordinate.Type], c)
	}

	return project.Project{
		Id:      "my-project",
		Configs: projectConfigs,
	}

}