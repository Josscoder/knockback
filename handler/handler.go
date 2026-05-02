package handler

import (
	"time"

	"github.com/josscoder/knockback/knockback"

	"github.com/df-mc/dragonfly/server/player"
	"github.com/df-mc/dragonfly/server/world"
)

type KnockBackHandler struct {
	player.NopHandler
}

func NewKnockBackHandler() *KnockBackHandler {
	return &KnockBackHandler{}
}

func (h *KnockBackHandler) HandleAttackEntity(_ *player.Context, _ world.Entity, force, height *float64, _ *bool) {
	cfg := knockback.GetKnockbackConfig()

	if force != nil {
		*force = cfg.HorizontalForce * cfg.Factor
	}
	if height != nil {
		vertical := cfg.VerticalForce * cfg.Factor
		if cfg.HeightLimiter > 0 && vertical > cfg.HeightLimiter {
			vertical = cfg.HeightLimiter
		}
		*height = vertical
	}
}

func (h *KnockBackHandler) HandleHurt(_ *player.Context, _ *float64, immune bool, attackImmunity *time.Duration, _ world.DamageSource) {
	if immune || attackImmunity == nil {
		return
	}
	cfg := knockback.GetKnockbackConfig()
	*attackImmunity = time.Duration(cfg.AttackCooldown) * time.Millisecond
}
