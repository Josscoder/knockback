package knockback

import "github.com/josscoder/knockback/config"

func Configure(settings Settings) {
	setManager(NewManager(settings))
}

func GetKnockbackConfig() *config.KnockbackConfig {
	return getManager().GetKnockbackConfig()
}

func CurrentPreset() string {
	return getManager().CurrentPreset()
}

func ListPresets() ([]string, error) {
	return getManager().ListPresets()
}

func PresetExists(name string) (bool, error) {
	return getManager().PresetExists(name)
}

func CreateOrUpdatePreset(name string, cfg *config.KnockbackConfig) error {
	return getManager().CreateOrUpdatePreset(name, cfg)
}

func SelectPreset(name string) error {
	return getManager().SelectPreset(name)
}

func DeletePreset(name string) error {
	return getManager().DeletePreset(name)
}

func LoadKnockbackConfig() error {
	return getManager().LoadKnockbackConfig()
}
