package knockback

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"sync"

	"knockback/config"
)

const (
	defaultPresetName = "default"
	activePresetFile  = ".active"
	presetDirName     = "presets"
)

type Settings struct {
	KnockbackPresetsPath string
}

type Manager struct {
	Settings Settings

	mu           sync.RWMutex
	kbConfig     *config.KnockbackConfig
	activePreset string
}

func NewManager(settings Settings) *Manager {
	return &Manager{
		Settings:     settings,
		kbConfig:     config.DefaultKnockbackConfig(),
		activePreset: defaultPresetName,
	}
}

func (m *Manager) GetKnockbackConfig() *config.KnockbackConfig {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.kbConfig.Clone()
}

func (m *Manager) CurrentPreset() string {
	m.mu.RLock()
	defer m.mu.RUnlock()
	if m.activePreset == "" {
		return defaultPresetName
	}
	return m.activePreset
}

func (m *Manager) ListPresets() ([]string, error) {
	dir, err := m.presetsDir()
	if err != nil {
		return nil, err
	}
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return nil, fmt.Errorf("failed creating presets directory %s: %w", dir, err)
	}

	entries, err := os.ReadDir(dir)
	if err != nil {
		return nil, fmt.Errorf("failed reading presets directory %s: %w", dir, err)
	}

	presets := make([]string, 0, len(entries))
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		name := entry.Name()
		if !strings.HasSuffix(name, ".json") {
			continue
		}
		presets = append(presets, strings.TrimSuffix(name, ".json"))
	}
	sort.Strings(presets)
	return presets, nil
}

func (m *Manager) PresetExists(name string) (bool, error) {
	normalized, err := normalizePresetName(name)
	if err != nil {
		return false, err
	}
	presetPath, err := m.presetPath(normalized)
	if err != nil {
		return false, err
	}
	_, err = os.Stat(presetPath)
	if err == nil {
		return true, nil
	}
	if errors.Is(err, os.ErrNotExist) {
		return false, nil
	}
	return false, err
}

func (m *Manager) CreateOrUpdatePreset(name string, cfg *config.KnockbackConfig) error {
	normalized, err := normalizePresetName(name)
	if err != nil {
		return err
	}
	if cfg == nil {
		return fmt.Errorf("knockback config cannot be nil")
	}
	if err := cfg.Validate(); err != nil {
		return err
	}
	if err := m.ensurePresetDir(); err != nil {
		return err
	}

	presetPath, err := m.presetPath(normalized)
	if err != nil {
		return err
	}
	if err := m.writeConfigFile(presetPath, cfg); err != nil {
		return err
	}

	m.mu.Lock()
	m.kbConfig = cfg.Clone()
	m.activePreset = normalized
	m.mu.Unlock()

	if err := m.writeActivePreset(normalized); err != nil {
		return err
	}
	return nil
}

func (m *Manager) SelectPreset(name string) error {
	normalized, err := normalizePresetName(name)
	if err != nil {
		return err
	}

	presetPath, err := m.presetPath(normalized)
	if err != nil {
		return err
	}
	cfg, err := m.readConfigFile(presetPath)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return fmt.Errorf("preset %q no existe", normalized)
		}
		return err
	}

	m.mu.Lock()
	m.kbConfig = cfg
	m.activePreset = normalized
	m.mu.Unlock()

	if err := m.writeActivePreset(normalized); err != nil {
		return err
	}
	return nil
}

func (m *Manager) DeletePreset(name string) error {
	normalized, err := normalizePresetName(name)
	if err != nil {
		return err
	}
	if normalized == defaultPresetName {
		return fmt.Errorf("no se puede eliminar el preset %q", defaultPresetName)
	}

	presetPath, err := m.presetPath(normalized)
	if err != nil {
		return err
	}
	if err := os.Remove(presetPath); err != nil {
		if !errors.Is(err, os.ErrNotExist) {
			return fmt.Errorf("failed deleting preset %q: %w", normalized, err)
		}
	}

	if m.CurrentPreset() == normalized {
		if err := m.SelectPreset(defaultPresetName); err != nil {
			if createErr := m.CreateOrUpdatePreset(defaultPresetName, config.DefaultKnockbackConfig()); createErr != nil {
				return fmt.Errorf("preset eliminado, pero no se pudo volver a %q: %w", defaultPresetName, err)
			}
		}
	}
	return nil
}

func (m *Manager) LoadKnockbackConfig() error {
	if err := m.ensurePresetDir(); err != nil {
		return err
	}

	active, err := m.readActivePreset()
	if err != nil {
		return err
	}

	activePath, err := m.presetPath(active)
	if err != nil {
		return err
	}
	cfg, err := m.readConfigFile(activePath)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			cfg = config.DefaultKnockbackConfig()
			if wErr := m.writeConfigFile(activePath, cfg); wErr != nil {
				return wErr
			}
		} else {
			return err
		}
	}

	m.mu.Lock()
	m.kbConfig = cfg
	m.activePreset = active
	m.mu.Unlock()

	if err := m.writeActivePreset(active); err != nil {
		return err
	}
	return nil
}

func (m *Manager) presetsDir() (string, error) {
	if m.Settings.KnockbackPresetsPath != "" {
		return m.Settings.KnockbackPresetsPath, nil
	}
	return presetDirName, nil
}

func (m *Manager) ensurePresetDir() error {
	dir, err := m.presetsDir()
	if err != nil {
		return err
	}
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return fmt.Errorf("failed creating presets directory %s: %w", dir, err)
	}
	return nil
}

func (m *Manager) activePresetPath() (string, error) {
	dir, err := m.presetsDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(dir, activePresetFile), nil
}

func (m *Manager) presetPath(name string) (string, error) {
	dir, err := m.presetsDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(dir, name+".json"), nil
}

func (m *Manager) readActivePreset() (string, error) {
	path, err := m.activePresetPath()
	if err != nil {
		return "", err
	}
	data, err := os.ReadFile(path)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return defaultPresetName, nil
		}
		return "", fmt.Errorf("failed reading active preset file %s: %w", path, err)
	}
	name := strings.TrimSpace(string(data))
	if name == "" {
		return defaultPresetName, nil
	}
	return normalizePresetName(name)
}

func (m *Manager) writeActivePreset(name string) error {
	path, err := m.activePresetPath()
	if err != nil {
		return err
	}
	if err := os.WriteFile(path, []byte(name), 0o644); err != nil {
		return fmt.Errorf("failed writing active preset file %s: %w", path, err)
	}
	return nil
}

func (m *Manager) readConfigFile(path string) (*config.KnockbackConfig, error) {
	raw, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	cfg := config.DefaultKnockbackConfig()
	if err := json.Unmarshal(raw, cfg); err != nil {
		return nil, fmt.Errorf("failed unmarshalling %s: %w", path, err)
	}
	if err := cfg.Validate(); err != nil {
		return nil, err
	}
	return cfg, nil
}

func (m *Manager) writeConfigFile(path string, cfg *config.KnockbackConfig) error {
	if cfg == nil {
		return fmt.Errorf("knockback config cannot be nil")
	}
	if err := cfg.Validate(); err != nil {
		return err
	}

	data, err := json.MarshalIndent(cfg, "", "  ")
	if err != nil {
		return fmt.Errorf("failed marshalling %s: %w", path, err)
	}

	dir := filepath.Dir(path)
	if dir != "." && dir != "" {
		if err := os.MkdirAll(dir, 0o755); err != nil {
			return fmt.Errorf("failed creating kb config directory %s: %w", dir, err)
		}
	}

	if err := os.WriteFile(path, data, 0o644); err != nil {
		return fmt.Errorf("failed writing %s: %w", path, err)
	}
	return nil
}

func normalizePresetName(name string) (string, error) {
	name = strings.TrimSpace(name)
	if name == "" {
		return "", fmt.Errorf("preset name cannot be empty")
	}
	for _, r := range name {
		if (r >= 'a' && r <= 'z') ||
			(r >= 'A' && r <= 'Z') ||
			(r >= '0' && r <= '9') ||
			r == '_' || r == '-' {
			continue
		}
		return "", fmt.Errorf("preset name %q has invalid characters", name)
	}
	return name, nil
}
