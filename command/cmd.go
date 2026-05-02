package command

import (
	"fmt"
	"strings"

	"knockback/config"
	"knockback/knockback"

	"github.com/df-mc/dragonfly/server/cmd"
	"github.com/df-mc/dragonfly/server/world"
	"github.com/sandertv/gophertunnel/minecraft/text"
)

func NewKnockbackCommand(manager *knockback.Manager) cmd.Command {
	return cmd.New(
		"kb",
		"Gestiona presets de knockback",
		[]string{"knockback"},
		kbHelp{Manager: manager},
		kbList{Manager: manager},
		kbCurrent{Manager: manager},
		kbSelect{Manager: manager},
		kbDelete{Manager: manager},
		kbCreate{Manager: manager},
		kbUpdate{Manager: manager},
	)
}

type kbHelp struct {
	Manager *knockback.Manager `cmd:"-"`
}

func (k kbHelp) Run(_ cmd.Source, output *cmd.Output, _ *world.Tx) {
	output.Print(text.Colourf("<yellow>Uso: /kb list | /kb current | /kb select <preset> | /kb create <preset> <h> <v> <cooldown> <limit> <factor> | /kb update <preset> <h> <v> <cooldown> <limit> <factor> | /kb delete <preset></yellow>"))
}

type kbList struct {
	Manager *knockback.Manager `cmd:"-"`
	Sub     cmd.SubCommand     `cmd:"list"`
}

func (k kbList) Run(_ cmd.Source, output *cmd.Output, _ *world.Tx) {
	if k.Manager == nil {
		output.Error("manager no configurado")
		return
	}
	presets, err := k.Manager.ListPresets()
	if err != nil {
		output.Error(err)
		return
	}
	if len(presets) == 0 {
		if err := k.Manager.CreateOrUpdatePreset("default", config.DefaultKnockbackConfig()); err != nil {
			output.Error(err)
			return
		}
		presets = []string{"default"}
	}
	output.Print(text.Colourf("<aqua>Preset activo:</aqua> <green>%s</green>", k.Manager.CurrentPreset()))
	output.Print(text.Colourf("<grey>Presets:</grey> <white>%s</white>", strings.Join(presets, ", ")))
}

type kbCurrent struct {
	Manager *knockback.Manager `cmd:"-"`
	Sub     cmd.SubCommand     `cmd:"current"`
}

func (k kbCurrent) Run(_ cmd.Source, output *cmd.Output, _ *world.Tx) {
	if k.Manager == nil {
		output.Error("manager no configurado")
		return
	}
	cfg := k.Manager.GetKnockbackConfig()
	output.Print(text.Colourf("<aqua>Preset activo:</aqua> <green>%s</green>", k.Manager.CurrentPreset()))
	output.Print(text.Colourf("<grey>%s</grey>", renderConfig(cfg)))
}

type kbSelect struct {
	Manager *knockback.Manager `cmd:"-"`
	Sub     cmd.SubCommand     `cmd:"select"`
	Name    string             `cmd:"preset"`
}

func (k kbSelect) Run(_ cmd.Source, output *cmd.Output, _ *world.Tx) {
	if k.Manager == nil {
		output.Error("manager no configurado")
		return
	}
	if err := k.Manager.SelectPreset(k.Name); err != nil {
		output.Error(err)
		return
	}
	output.Print(text.Colourf("<green>Preset seleccionado:</green> <white>%s</white>", k.Name))
	output.Print(text.Colourf("<grey>%s</grey>", renderConfig(k.Manager.GetKnockbackConfig())))
}

type kbDelete struct {
	Manager *knockback.Manager `cmd:"-"`
	Sub     cmd.SubCommand     `cmd:"delete"`
	Name    string             `cmd:"preset"`
}

func (k kbDelete) Run(_ cmd.Source, output *cmd.Output, _ *world.Tx) {
	if k.Manager == nil {
		output.Error("manager no configurado")
		return
	}
	if err := k.Manager.DeletePreset(k.Name); err != nil {
		output.Error(err)
		return
	}
	output.Print(text.Colourf("<red>Preset eliminado:</red> <white>%s</white>", k.Name))
	output.Print(text.Colourf("<aqua>Preset activo actual:</aqua> <green>%s</green>", k.Manager.CurrentPreset()))
}

type kbCreate struct {
	Manager         *knockback.Manager `cmd:"-"`
	Sub             cmd.SubCommand     `cmd:"create"`
	Name            string             `cmd:"preset"`
	HorizontalForce float64            `cmd:"horizontal"`
	VerticalForce   float64            `cmd:"vertical"`
	AttackCooldown  float64            `cmd:"cooldown"`
	HeightLimiter   float64            `cmd:"height_limiter"`
	Factor          float64            `cmd:"factor"`
}

func (k kbCreate) Run(_ cmd.Source, output *cmd.Output, _ *world.Tx) {
	if k.Manager == nil {
		output.Error("manager no configurado")
		return
	}
	exists, err := k.Manager.PresetExists(k.Name)
	if err != nil {
		output.Error(err)
		return
	}
	if exists {
		output.Errorf("el preset %q ya existe", k.Name)
		return
	}
	cfg, err := k.buildConfig()
	if err != nil {
		output.Error(err)
		return
	}
	if err := k.Manager.CreateOrUpdatePreset(k.Name, cfg); err != nil {
		output.Error(err)
		return
	}
	output.Print(text.Colourf("<green>Preset creado y seleccionado:</green> <white>%s</white>", k.Name))
	output.Print(text.Colourf("<grey>%s</grey>", renderConfig(cfg)))
}

type kbUpdate struct {
	Manager         *knockback.Manager `cmd:"-"`
	Sub             cmd.SubCommand     `cmd:"update"`
	Name            string             `cmd:"preset"`
	HorizontalForce float64            `cmd:"horizontal"`
	VerticalForce   float64            `cmd:"vertical"`
	AttackCooldown  float64            `cmd:"cooldown"`
	HeightLimiter   float64            `cmd:"height_limiter"`
	Factor          float64            `cmd:"factor"`
}

func (k kbUpdate) Run(_ cmd.Source, output *cmd.Output, _ *world.Tx) {
	if k.Manager == nil {
		output.Error("manager no configurado")
		return
	}
	exists, err := k.Manager.PresetExists(k.Name)
	if err != nil {
		output.Error(err)
		return
	}
	if !exists {
		output.Errorf("el preset %q no existe", k.Name)
		return
	}
	cfg, err := k.buildConfig()
	if err != nil {
		output.Error(err)
		return
	}
	if err := k.Manager.CreateOrUpdatePreset(k.Name, cfg); err != nil {
		output.Error(err)
		return
	}
	output.Print(text.Colourf("<green>Preset actualizado y seleccionado:</green> <white>%s</white>", k.Name))
	output.Print(text.Colourf("<grey>%s</grey>", renderConfig(cfg)))
}

func (k kbCreate) buildConfig() (*config.KnockbackConfig, error) {
	return buildConfig(k.HorizontalForce, k.VerticalForce, k.AttackCooldown, k.HeightLimiter, k.Factor)
}

func (k kbUpdate) buildConfig() (*config.KnockbackConfig, error) {
	return buildConfig(k.HorizontalForce, k.VerticalForce, k.AttackCooldown, k.HeightLimiter, k.Factor)
}

func buildConfig(horizontal, vertical, cooldown, limiter, factor float64) (*config.KnockbackConfig, error) {
	cfg := &config.KnockbackConfig{
		HorizontalForce: horizontal,
		VerticalForce:   vertical,
		AttackCooldown:  cooldown,
		HeightLimiter:   limiter,
		Factor:          factor,
	}
	if err := cfg.Validate(); err != nil {
		return nil, err
	}
	return cfg, nil
}

func renderConfig(cfg *config.KnockbackConfig) string {
	if cfg == nil {
		return "config: <nil>"
	}
	return fmt.Sprintf(
		"horizontal=%.3f vertical=%.3f cooldown=%.3f height_limiter=%.3f factor=%.3f",
		cfg.HorizontalForce,
		cfg.VerticalForce,
		cfg.AttackCooldown,
		cfg.HeightLimiter,
		cfg.Factor,
	)
}
