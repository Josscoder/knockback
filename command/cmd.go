package command

import (
	"fmt"
	"strconv"
	"strings"

	"knockback/config"
	"knockback/knockback"

	"github.com/df-mc/dragonfly/server/cmd"
	"github.com/df-mc/dragonfly/server/player"
	"github.com/df-mc/dragonfly/server/world"
	"github.com/sandertv/gophertunnel/minecraft/text"
	form "github.com/twistedasylummc/inline-forms"
)

func NewKnockbackCommand(manager *knockback.Manager) cmd.Command {
	return cmd.New(
		"kb",
		"Gestiona presets de knockback",
		[]string{"knockback"},
		KnockbackCommand{Manager: manager},
	)
}

type KnockbackCommand struct {
	Manager *knockback.Manager `cmd:"-"`
}

func (k KnockbackCommand) Run(source cmd.Source, output *cmd.Output, _ *world.Tx) {
	if k.Manager == nil {
		output.Error("manager no configurado")
		return
	}
	p, ok := source.(*player.Player)
	if !ok {
		output.Error("este comando solo puede usarse en juego")
		return
	}
	openMainMenu(p, k.Manager)
}

func openMainMenu(p *player.Player, manager *knockback.Manager) {
	current := manager.CurrentPreset()

	menu := &form.Menu{
		Title:   "Knockback presets",
		Content: text.Colourf("<aqua>Preset activo:</aqua> <green>%s</green>", current),
		Elements: []form.MenuElement{
			form.Button{Text: text.Colourf("<green>Ver preset activo</green>"), Submit: func(tx *world.Tx) {
				openCurrentPreset(p, manager)
			}},
			form.Button{Text: text.Colourf("<yellow>Seleccionar preset</yellow>"), Submit: func(tx *world.Tx) {
				openSelectPresetForm(p, manager)
			}},
			form.Button{Text: text.Colourf("<green>Crear preset</green>"), Submit: func(tx *world.Tx) {
				openCreatePresetForm(p, manager)
			}},
			form.Button{Text: text.Colourf("<aqua>Editar preset activo</aqua>"), Submit: func(tx *world.Tx) {
				openEditCurrentPresetForm(p, manager)
			}},
			form.Button{Text: text.Colourf("<red>Eliminar preset</red>"), Submit: func(tx *world.Tx) {
				openDeletePresetForm(p, manager)
			}},
		},
	}
	p.SendForm(menu)
}

func openCurrentPreset(p *player.Player, manager *knockback.Manager) {
	cfg := manager.GetKnockbackConfig()
	openNotice(
		p,
		manager,
		"Preset activo",
		text.Colourf("<green>%s</green>\n<grey>%s</grey>", manager.CurrentPreset(), renderConfig(cfg)),
	)
}

func openSelectPresetForm(p *player.Player, manager *knockback.Manager) {
	presets, err := ensurePresets(manager)
	if err != nil {
		openError(p, manager, err)
		return
	}

	current := manager.CurrentPreset()
	defaultIndex := indexOf(presets, current)
	if defaultIndex < 0 {
		defaultIndex = 0
	}
	selected := presets[defaultIndex]

	custom := &form.Custom{
		Title: "Seleccionar preset",
		Elements: []form.Element{
			form.Label{Text: "Selecciona el preset que quieres activar."},
			form.Dropdown{
				Text:         "Preset",
				Options:      presets,
				DefaultIndex: defaultIndex,
				Submit: func(index int, option string) {
					selected = option
				},
			},
		},
		Submit: func(closed bool, _ []any, tx *world.Tx) {
			if closed {
				openMainMenu(p, manager)
				return
			}
			if err := manager.SelectPreset(selected); err != nil {
				openError(p, manager, err)
				return
			}
			openNotice(
				p,
				manager,
				"Preset seleccionado",
				text.Colourf("<green>%s</green>\n<grey>%s</grey>", selected, renderConfig(manager.GetKnockbackConfig())),
			)
		},
	}
	p.SendForm(custom)
}

func openCreatePresetForm(p *player.Player, manager *knockback.Manager) {
	cfg := manager.GetKnockbackConfig()

	var name, horizontal, vertical, cooldown, limiter, factor string
	custom := &form.Custom{
		Title: "Crear preset",
		Elements: []form.Element{
			form.Label{Text: "Crea un nuevo preset de knockback."},
			form.Input{Text: "Nombre del preset", Placeholder: "combo", Default: "", Submit: func(v string) { name = v }},
			form.Input{Text: "Horizontal force", Placeholder: "0.4", Default: formatFloat(cfg.HorizontalForce), Submit: func(v string) { horizontal = v }},
			form.Input{Text: "Vertical force", Placeholder: "0.4", Default: formatFloat(cfg.VerticalForce), Submit: func(v string) { vertical = v }},
			form.Input{Text: "Attack cooldown (ms)", Placeholder: "100", Default: formatInt(cfg.AttackCooldown), Submit: func(v string) { cooldown = v }},
			form.Input{Text: "Height limiter", Placeholder: "0.4", Default: formatFloat(cfg.HeightLimiter), Submit: func(v string) { limiter = v }},
			form.Input{Text: "Factor", Placeholder: "1", Default: formatFloat(cfg.Factor), Submit: func(v string) { factor = v }},
		},
		Submit: func(closed bool, _ []any, tx *world.Tx) {
			if closed {
				openMainMenu(p, manager)
				return
			}

			name = strings.TrimSpace(name)
			if name == "" {
				openError(p, manager, fmt.Errorf("el nombre del preset no puede estar vacío"))
				return
			}

			exists, err := manager.PresetExists(name)
			if err != nil {
				openError(p, manager, err)
				return
			}
			if exists {
				openError(p, manager, fmt.Errorf("el preset %q ya existe", name))
				return
			}

			created, err := buildConfigFromStrings(horizontal, vertical, cooldown, limiter, factor)
			if err != nil {
				openError(p, manager, err)
				return
			}
			if err := manager.CreateOrUpdatePreset(name, created); err != nil {
				openError(p, manager, err)
				return
			}
			openNotice(
				p,
				manager,
				"Preset creado",
				text.Colourf("<green>%s</green>\n<grey>%s</grey>", name, renderConfig(created)),
			)
		},
	}
	p.SendForm(custom)
}

func openEditCurrentPresetForm(p *player.Player, manager *knockback.Manager) {
	name := manager.CurrentPreset()
	cfg := manager.GetKnockbackConfig()

	var horizontal, vertical, cooldown, limiter, factor string
	custom := &form.Custom{
		Title: "Editar preset activo",
		Elements: []form.Element{
			form.Label{Text: text.Colourf("Editando preset: <green>%s</green>", name)},
			form.Input{Text: "Horizontal force", Placeholder: "0.4", Default: formatFloat(cfg.HorizontalForce), Submit: func(v string) { horizontal = v }},
			form.Input{Text: "Vertical force", Placeholder: "0.4", Default: formatFloat(cfg.VerticalForce), Submit: func(v string) { vertical = v }},
			form.Input{Text: "Attack cooldown (ms)", Placeholder: "100", Default: formatInt(cfg.AttackCooldown), Submit: func(v string) { cooldown = v }},
			form.Input{Text: "Height limiter", Placeholder: "0.4", Default: formatFloat(cfg.HeightLimiter), Submit: func(v string) { limiter = v }},
			form.Input{Text: "Factor", Placeholder: "1", Default: formatFloat(cfg.Factor), Submit: func(v string) { factor = v }},
		},
		Submit: func(closed bool, _ []any, tx *world.Tx) {
			if closed {
				openMainMenu(p, manager)
				return
			}
			updated, err := buildConfigFromStrings(horizontal, vertical, cooldown, limiter, factor)
			if err != nil {
				openError(p, manager, err)
				return
			}
			if err := manager.CreateOrUpdatePreset(name, updated); err != nil {
				openError(p, manager, err)
				return
			}
			openNotice(
				p,
				manager,
				"Preset actualizado",
				text.Colourf("<green>%s</green>\n<grey>%s</grey>", name, renderConfig(updated)),
			)
		},
	}
	p.SendForm(custom)
}

func openDeletePresetForm(p *player.Player, manager *knockback.Manager) {
	presets, err := ensurePresets(manager)
	if err != nil {
		openError(p, manager, err)
		return
	}

	deletable := make([]string, 0, len(presets))
	for _, preset := range presets {
		if preset != "default" {
			deletable = append(deletable, preset)
		}
	}
	if len(deletable) == 0 {
		openNotice(p, manager, "Eliminar preset", "<yellow>No hay presets eliminables.</yellow>")
		return
	}

	selected := deletable[0]
	custom := &form.Custom{
		Title: "Eliminar preset",
		Elements: []form.Element{
			form.Label{Text: "Selecciona el preset que quieres eliminar."},
			form.Dropdown{
				Text:         "Preset",
				Options:      deletable,
				DefaultIndex: 0,
				Submit: func(index int, option string) {
					selected = option
				},
			},
		},
		Submit: func(closed bool, _ []any, tx *world.Tx) {
			if closed {
				openMainMenu(p, manager)
				return
			}
			openDeleteConfirmModal(p, manager, selected)
		},
	}
	p.SendForm(custom)
}

func openDeleteConfirmModal(p *player.Player, manager *knockback.Manager, name string) {
	modal := &form.Modal{
		Title:   "Confirmar eliminación",
		Content: text.Colourf("¿Seguro que quieres eliminar <red>%s</red>?", name),
		Button1: form.Button{
			Text: "Eliminar",
			Submit: func(tx *world.Tx) {
				if err := manager.DeletePreset(name); err != nil {
					openError(p, manager, err)
					return
				}
				openNotice(
					p,
					manager,
					"Preset eliminado",
					text.Colourf("<red>%s</red>\n<aqua>Activo actual:</aqua> <green>%s</green>", name, manager.CurrentPreset()),
				)
			},
		},
		Button2: form.Button{
			Text: "Cancelar",
			Submit: func(tx *world.Tx) {
				openMainMenu(p, manager)
			},
		},
		Submit: func(closed bool, tx *world.Tx) {
			if closed {
				openMainMenu(p, manager)
			}
		},
	}
	p.SendForm(modal)
}

func openNotice(p *player.Player, manager *knockback.Manager, title, content string) {
	menu := &form.Menu{
		Title:   title,
		Content: content,
		Elements: []form.MenuElement{
			form.Button{Text: "Volver", Submit: func(tx *world.Tx) { openMainMenu(p, manager) }},
		},
	}
	p.SendForm(menu)
}

func openError(p *player.Player, manager *knockback.Manager, err error) {
	openNotice(
		p,
		manager,
		"Error",
		text.Colourf("<red>%v</red>", err),
	)
}

func ensurePresets(manager *knockback.Manager) ([]string, error) {
	presets, err := manager.ListPresets()
	if err != nil {
		return nil, err
	}
	if len(presets) == 0 {
		if err := manager.CreateOrUpdatePreset("default", config.DefaultKnockbackConfig()); err != nil {
			return nil, err
		}
		return []string{"default"}, nil
	}
	return presets, nil
}

func indexOf(values []string, target string) int {
	for i, value := range values {
		if value == target {
			return i
		}
	}
	return -1
}

func buildConfigFromStrings(horizontal, vertical, cooldown, limiter, factor string) (*config.KnockbackConfig, error) {
	h, err := parseNonNegativeFloat(horizontal, "horizontal force")
	if err != nil {
		return nil, err
	}
	v, err := parseNonNegativeFloat(vertical, "vertical force")
	if err != nil {
		return nil, err
	}
	cooldownMS, err := parseNonNegativeInt(cooldown, "attack cooldown")
	if err != nil {
		return nil, err
	}
	l, err := parseNonNegativeFloat(limiter, "height limiter")
	if err != nil {
		return nil, err
	}
	f, err := parseNonNegativeFloat(factor, "factor")
	if err != nil {
		return nil, err
	}

	cfg := &config.KnockbackConfig{
		HorizontalForce: h,
		VerticalForce:   v,
		AttackCooldown:  cooldownMS,
		HeightLimiter:   l,
		Factor:          f,
	}
	if err := cfg.Validate(); err != nil {
		return nil, err
	}
	return cfg, nil
}

func parseNonNegativeFloat(raw string, field string) (float64, error) {
	value, err := strconv.ParseFloat(strings.TrimSpace(raw), 64)
	if err != nil {
		return 0, fmt.Errorf("%s debe ser un número válido", field)
	}
	if value < 0 {
		return 0, fmt.Errorf("%s no puede ser negativo", field)
	}
	return value, nil
}

func formatFloat(v float64) string {
	return strconv.FormatFloat(v, 'f', -1, 64)
}

func formatInt(v int64) string {
	return strconv.FormatInt(v, 10)
}

func parseNonNegativeInt(raw string, field string) (int64, error) {
	value, err := strconv.ParseInt(strings.TrimSpace(raw), 10, 64)
	if err != nil {
		return 0, fmt.Errorf("%s debe ser un número entero válido", field)
	}
	if value < 0 {
		return 0, fmt.Errorf("%s no puede ser negativo", field)
	}
	return value, nil
}

func renderConfig(cfg *config.KnockbackConfig) string {
	if cfg == nil {
		return "config: <nil>"
	}
	return fmt.Sprintf(
		"horizontal=%.3f vertical=%.3f cooldown=%dms height_limiter=%.3f factor=%.3f",
		cfg.HorizontalForce,
		cfg.VerticalForce,
		cfg.AttackCooldown,
		cfg.HeightLimiter,
		cfg.Factor,
	)
}
