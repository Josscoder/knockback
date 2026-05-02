# Knockback

This project applies and manages knockback presets for Dragonfly-based servers.

## Using it outside this plugin

You can reuse the logic as a package in another Go/Dragonfly project using these packages:

- `knockback/config` (config structure and validation)
- `knockback/knockback` (disk-based preset manager)
- `knockback/handler` (handler that applies knockback and cooldown)

### 1) Initialize the manager

```go
kbManager := knockback.NewManager(knockback.Settings{
	KnockbackPresetsPath: "presets",
})
if err := kbManager.LoadKnockbackConfig(); err != nil {
	panic(err)
}
```

### 2) Register the handler for players

```go
for p := range srv.Accept() {
	p.Handle(handler.NewKnockBackHandler(kbManager))
}
```

### 3) Create/edit/select presets in code

```go
cfg := &config.KnockbackConfig{
	HorizontalForce: 0.35,
	VerticalForce:   0.35,
	AttackCooldown:  120,
	HeightLimiter:   0.4,
	Factor:          1.0,
}

if err := kbManager.CreateOrUpdatePreset("ranked", cfg); err != nil {
	panic(err)
}

if err := kbManager.SelectPreset("ranked"); err != nil {
	panic(err)
}
```

## Preset format (`*.json`)

```json
{
  "horizontal_force": 0.2,
  "vertical_force": 0.2,
  "attack_cooldown": 300,
  "height_limiter": 0.3,
  "factor": 1
}
```

## File structure

- `presets/.active`: active preset name.
- `presets/<name>.json`: each preset configuration.

If no active preset exists, `default` is used.

## Values and rules

- All values must be `>= 0`.
- `attack_cooldown` is in milliseconds.
- Preset names only allow letters, numbers, `_`, and `-`.

## Run this project as a reference

```bash
go run .
```

In-game, use `/kb` to open the preset GUI menu.
