package config

import "fmt"

type KnockbackConfig struct {
	HorizontalForce float64 `json:"horizontal_force"`
	VerticalForce   float64 `json:"vertical_force"`
	AttackCooldown  int64   `json:"attack_cooldown"`
	HeightLimiter   float64 `json:"height_limiter"`
	Factor          float64 `json:"factor"`
}

func DefaultKnockbackConfig() *KnockbackConfig {
	return &KnockbackConfig{
		HorizontalForce: 0.4,
		VerticalForce:   0.4,
		AttackCooldown:  0,
		HeightLimiter:   0.4,
		Factor:          1,
	}
}

func (k *KnockbackConfig) Clone() *KnockbackConfig {
	if k == nil {
		return DefaultKnockbackConfig()
	}
	cloned := *k
	return &cloned
}

func (k *KnockbackConfig) Validate() error {
	if k == nil {
		return fmt.Errorf("knockback config cannot be nil")
	}
	if k.HorizontalForce < 0 {
		return fmt.Errorf("horizontal_force cannot be negative")
	}
	if k.VerticalForce < 0 {
		return fmt.Errorf("vertical_force cannot be negative")
	}
	if k.AttackCooldown < 0 {
		return fmt.Errorf("attack_cooldown cannot be negative")
	}
	if k.HeightLimiter < 0 {
		return fmt.Errorf("height_limiter cannot be negative")
	}
	if k.Factor < 0 {
		return fmt.Errorf("factor cannot be negative")
	}
	return nil
}
