package main

import (
	"log/slog"

	"github.com/josscoder/knockback/command"
	"github.com/josscoder/knockback/handler"
	"github.com/josscoder/knockback/knockback"

	"github.com/df-mc/dragonfly/server"
	"github.com/df-mc/dragonfly/server/cmd"
	"github.com/df-mc/dragonfly/server/player/chat"
	"github.com/df-mc/dragonfly/server/world"
)

func main() {
	slog.SetLogLoggerLevel(slog.LevelDebug)
	chat.Global.Subscribe(chat.StdoutSubscriber{})

	if err := knockback.LoadKnockbackConfig(); err != nil {
		panic(err)
	}
	cmd.Register(command.NewKnockbackCommand())

	srvConf := server.DefaultConfig()
	srvConf.Players.SaveData = false

	conf, err := srvConf.Config(slog.Default())
	if err != nil {
		panic(err)
	}

	srv := conf.New()
	srv.CloseOnProgramEnd()

	srv.Listen()
	for p := range srv.Accept() {
		p.Handle(handler.NewKnockBackHandler())
		p.SetGameMode(world.GameModeSurvival)
	}
}
