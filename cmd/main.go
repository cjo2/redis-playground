package main

import (
	"context"
	"crypto/sha1"
	"fmt"
	"log/slog"

	"github.com/redis/rueidis"
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

	// basic script
	script := "local key = KEYS[1]\nlocal increment_value = ARGV[1]\n\nreturn redis.call('INCRBY', key, increment_value)"

	// create sha1 hash to check if the script exists
	h := sha1.New()
	h.Write([]byte(script))
	sha1Res := h.Sum(nil)
	hash := fmt.Sprintf("%x", sha1Res)

	ReportScriptExists(context.TODO(), client, hash)

	redisScript := rueidis.NewLuaScript(script)
	if err := redisScript.Exec(context.TODO(), client, []string{"test"}, []string{"1"}).Error(); err != nil {
		slog.With("error", err.Error()).Error("failed to execute basic script")
		return
	}

	ReportScriptExists(context.TODO(), client, hash)

	// sending whole JSON (string-ified) over the network...
	script = "local key = KEYS[1]\nlocal json_value = ARGV[1]\n\nreturn redis.call('JSON.SET', key, '$', json_value)"

	type Thing struct {
		Name  string `json:"name,omitempty"`
		Value string `json:"value,omitempty"`
	}

	thing := Thing{
		Name:  "name of thing",
		Value: "value of thing",
	}

	thingJson := rueidis.JSON(thing)

	redisScript = rueidis.NewLuaScript(script)

	if err := redisScript.Exec(context.TODO(), client, []string{"jsontest"}, []string{thingJson}).Error(); err != nil {
		slog.With("error", err.Error()).Error("failed to execute json script")
	}
}

func ReportScriptExists(ctx context.Context, client rueidis.Client, hash string) {
	if ok, err := scriptExists(ctx, client, hash); err != nil {
		slog.With("error", err.Error()).Error("SCRIPT EXISTS failed")
	} else if ok {
		slog.Info("script exists in Redis")
	} else {
		slog.Info("script does not exist in Redis")
	}
}

func scriptExists(ctx context.Context, client rueidis.Client, hash string) (bool, error) {
	res, err := client.Do(ctx, client.B().ScriptExists().Sha1(hash).Build()).AsIntSlice()
	if err != nil {
		return false, err
	}

	return res[0] == 1, nil
}
