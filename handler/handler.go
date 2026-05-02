package handler

import (
	"sync"
	"time"

	"knockback/knockback"

	"github.com/df-mc/dragonfly/server/player"
	"github.com/df-mc/dragonfly/server/world"
)

type KnockBackHandler struct {
	player.NopHandler

	Manager *knockback.Manager

	mu         sync.Mutex
	lastAttack time.Time
}

func NewKnockBackHandler(manager *knockback.Manager) *KnockBackHandler {
	return &KnockBackHandler{Manager: manager}
}

func (h *KnockBackHandler) HandleAttackEntity(_ *player.Context, _ world.Entity, force, height *float64, _ *bool) {
	if h.Manager == nil {
		return
	}
	cfg := h.Manager.GetKnockbackConfig()

	multiplier := cfg.Factor
	if cfg.AttackCooldown > 0 {
		now := time.Now()
		h.mu.Lock()
		if !h.lastAttack.IsZero() {
			elapsed := now.Sub(h.lastAttack).Seconds()
			if elapsed < cfg.AttackCooldown {
				multiplier *= elapsed / cfg.AttackCooldown
			}
		}
		h.lastAttack = now
		h.mu.Unlock()
	}

	if force != nil {
		*force = cfg.HorizontalForce * multiplier
	}
	if height != nil {
		vertical := cfg.VerticalForce * multiplier
		if cfg.HeightLimiter > 0 && vertical > cfg.HeightLimiter {
			vertical = cfg.HeightLimiter
		}
		*height = vertical
	}
}
