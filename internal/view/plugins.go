package view

import (
	"bytes"
	"fmt"
	"html/template"

	"github.com/antvirf/stui/internal/config"
	"github.com/antvirf/stui/internal/logger"
	"github.com/gdamore/tcell/v2"
)

const (
	NO_KEY_FOUND = -1
)

// Exec a plugin command when on a particular page
func (a *App) ExecutePluginForShortcut(key tcell.Key, page string, rowId string) {
	availablePlugins := getPluginsForPage(page)
	for _, plugin := range availablePlugins {
		parsedKey := asKey(plugin.Shortcut)
		if parsedKey == NO_KEY_FOUND {
			logger.Debugf("plugin command %s has invalid shortcut: '%s'", plugin.Name, plugin.Shortcut)
			continue // The key is invalid so we skip this plugin
		}

		if parsedKey == key {
			provider := a.GetProviderForPage(page)
			if provider == nil {
				break
			}

			// Get row data for this row id
			rowData, err := provider.Data().GetRowAsMapById(rowId)
			if err != nil {
				logger.Printf("could not get data for this row")
				break
			}

			parsedCommand := a.ParsePluginCommand(plugin.Command, rowData, page)

			a.ShowCommandModal(parsedCommand, page, plugin.ExecuteImmediately, plugin.ClosePromptAfterExecute)

			// Stop processing further plugins - first one takes precedence.
			break
		}
	}
}

func (a *App) ParsePluginCommand(command string, data map[string]string, page string) string {
	tmpl, err := template.New("command").Parse(command)
	if err != nil {
		return fmt.Sprintf("invalid template: %v", err)
	}
	var output bytes.Buffer
	err = tmpl.Execute(&output, data)
	if err != nil {
		return fmt.Sprintf("failed to render template: %v", err)
	}
	return output.String()
}

// AsKey maps a string representation of a key to a tcell key.
func asKey(key string) tcell.Key {
	for k, v := range tcell.KeyNames {
		if key == v {
			return k
		}
	}
	return NO_KEY_FOUND
}

func getPluginsForPage(page string) []config.PluginConfig {
	plugins := []config.PluginConfig{}
	for _, plugin := range config.ConfigFile.Plugins {
		if plugin.ActivePage == page {
			plugins = append(plugins, plugin)

		}
	}
	return plugins
}

// Update helper keybinds
func GetKeyboardShortcutHelperForPage(page string) string {
	plugins := getPluginsForPage(page)
	if len(plugins) == 0 {
		return ""
	}
	helper := "CUSTOM PLUGIN SHORTCUTS (in current view only)"

	for _, plugin := range plugins {

		// Figure out the nice print format for the key
		// If it's missing, this is where we can inform the user.
		keyString := plugin.Shortcut
		if asKey(plugin.Shortcut) == NO_KEY_FOUND {
			keyString = "(N/A)"
		}

		helper = fmt.Sprintf(
			"%s\n%s",
			helper,
			fmt.Sprintf("%-9s%s (%s)", keyString, plugin.Name, plugin.Command),
		)

	}
	return helper
}
