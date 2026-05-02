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

func (h *KnockBackHandler) HandleAttackEntity(ctx *player.Context, target world.Entity, force, height *float64, _ *bool) {
	cfg := knockback.GetKnockbackConfig()

	if force != nil {
		*force = cfg.HorizontalForce
	}

	if height != nil {
		verticalForce := cfg.VerticalForce

		if cfg.HeightLimiter > 0 {
			dist := target.Position().Y() - ctx.Val().Position().Y()
			maxLayers := float64(cfg.HeightLimiter)
			if dist+verticalForce > maxLayers {
				verticalForce = maxLayers - dist
				if verticalForce < 0 {
					verticalForce = 0
				}
			}
		}

		*height = verticalForce
	}
}

func (h *KnockBackHandler) HandleHurt(_ *player.Context, _ *float64, immune bool, attackImmunity *time.Duration, _ world.DamageSource) {
	if immune || attackImmunity == nil {
		return
	}
	cfg := knockback.GetKnockbackConfig()
	*attackImmunity = time.Duration(cfg.AttackCooldown) * time.Millisecond
}
