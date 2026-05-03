# Knockback

This project applies and manages knockback presets for Dragonfly-based servers.

## Using it outside this plugin

You can reuse the logic as a package in another Go/Dragonfly project using these packages:

- `github.com/josscoder/knockback/config` (config structure and validation)
- `github.com/josscoder/knockback/knockback` (disk-based preset manager)
- `github.com/josscoder/knockback/handler` (handler that applies knockback and cooldown)

### 1) Initialize/load the package API

```go
knockback.Configure(knockback.Settings{
	KnockbackPresetsPath: "presets",
})
if err := knockback.LoadKnockbackConfig(); err != nil {
	panic(err)
}
```

### 2) Register the handler for players

```go
for p := range srv.Accept() {
	p.Handle(handler.NewKnockBackHandler())
}
```

### 3) Create/edit/select presets in code

```go
cfg := &config.KnockbackConfig{
	HorizontalForce: 0.35,
	VerticalForce:   0.35,
	AttackCooldown:  0.12,
	HeightLimiter:   1,
}

if err := knockback.CreateOrUpdatePreset("ranked", cfg); err != nil {
	panic(err)
}

if err := knockback.SelectPreset("ranked"); err != nil {
	panic(err)
}
```

## Preset format (`*.json`)

```json
{
  "horizontal_force": 0.2,
  "vertical_force": 0.2,
  "attack_cooldown": 0.3,
  "height_limiter": 1
}
```

## File structure

- `presets/.active`: active preset name.
- `presets/<name>.json`: each preset configuration.

If no active preset exists, `default` is used.

## Values and rules

- All values must be `>= 0`.
- `attack_cooldown` is in seconds (supports decimals, e.g. `1.5`).
- `height_limiter` is an integer (Y layers).
- Preset names only allow letters, numbers, `_`, and `-`.

## Run this project as a reference

```bash
go run .
```

In-game, use `/kb` to open the preset GUI menu.
