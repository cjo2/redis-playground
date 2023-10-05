package main

import (
	"context"
	"crypto/sha1"
	"fmt"
	"github.com/redis/rueidis"
	"log/slog"
)

func main() {
	var client rueidis.Client

	if c, err := rueidis.NewClient(rueidis.ClientOption{
		InitAddress: []string{"localhost:6379"},
		Username:    "",
		Password:    "",
	}); err != nil {
		slog.With("error", err).Error("could not create client")
		panic("could not create client")
	} else {
		client = c
	}

	script := "local key = KEYS[1]\nlocal increment_value = ARGV[1]\n\nreturn redis.call('INCRBY', key, increment_value)"
	h := sha1.New()
	h.Write([]byte(script))
	sha1Res := h.Sum(nil)
	hash := fmt.Sprintf("%x", sha1Res)

	slog.With("result", hash).Info("created sha1 of script")

	asBool, err := client.Do(context.TODO(), client.B().ScriptExists().Sha1(hash).Build()).AsIntSlice()
	if err != nil {
		return
	}

	slog.With("exists", asBool).Info("executed SCRIPT EXISTS")

	l := rueidis.NewLuaScript(script)
	l.Exec(context.TODO(), client, []string{"test"}, []string{"1"})
}
