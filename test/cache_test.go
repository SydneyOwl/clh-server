package test

import (
	"crypto/rand"
	"github.com/sydneyowl/clh-server/internal/cache"
	"github.com/sydneyowl/clh-server/msgproto"
	"github.com/sydneyowl/clh-server/pkg/msg"
	"strconv"
	"testing"
	"time"
)

func TestCache(t *testing.T) {
	tm := cache.NewMemoryCache()

	for i := range 10 {
		err := tm.Add("TEST", &msgproto.WsjtxMessage{
			Payload: &msgproto.WsjtxMessage_Decode{
				Decode: &msgproto.Decode{
					New:               false,
					Time:              0,
					Snr:               0,
					OffsetTimeSeconds: 0,
					OffsetFrequencyHz: 0,
					Mode:              "FT8",
					Message:           "InfoTest" + strconv.Itoa(i),
					LowConfidence:     false,
					OffAir:            false,
				},
			},
			Timestamp: time.Now().Unix(),
		})
		if err != nil {
			t.Fatal(err)
		}
	}

	all, err := tm.ReadAll("TEST")
	if err != nil {
		t.Fatal(err)
	}

	for _, v := range all {
		t.Logf("cache content: %s", v)
	}

	all, err = tm.ReadAll("TEST")
	if err != nil {
		t.Fatal(err)
	}
	t.Logf("%s", strconv.Itoa(len(all)))

	go func() {
		for range 100 {
			tm.Add("TEST", &msgproto.WsjtxMessage{
				Payload: &msgproto.WsjtxMessage_Decode{
					Decode: &msgproto.Decode{
						New:               false,
						Time:              0,
						Snr:               0,
						OffsetTimeSeconds: 0,
						OffsetFrequencyHz: 0,
						Mode:              "FT8",
						Message:           "InfoTestFF" + rand.Text(),
						LowConfidence:     false,
						OffAir:            false,
					},
				},
				Timestamp: time.Now().Unix(),
			})
			time.Sleep(1 * time.Second)
		}
	}()
	err = tm.ReadUntil("TEST", func(message msg.Message) {
		t.Logf("We got %s", message)
	}, make(<-chan struct{}))
	if err != nil {
		t.Fatal(err)
	}
	select {}
}
