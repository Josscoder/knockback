package handler

import (
	"time"

	"github.com/josscoder/knockback/knockback"

	"github.com/df-mc/dragonfly/server/player"
	"github.com/df-mc/dragonfly/server/world"
)

type KnockBackHandler struct {
	player.NopHandler

	Manager *knockback.Manager
}

func NewKnockBackHandler(manager *knockback.Manager) *KnockBackHandler {
	return &KnockBackHandler{Manager: manager}
}

func (h *KnockBackHandler) HandleAttackEntity(_ *player.Context, _ world.Entity, force, height *float64, _ *bool) {
	if h.Manager == nil {
		return
	}
	cfg := h.Manager.GetKnockbackConfig()

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
	if immune || h.Manager == nil || attackImmunity == nil {
		return
	}
	cfg := h.Manager.GetKnockbackConfig()
	*attackImmunity = time.Duration(cfg.AttackCooldown) * time.Millisecond
}
