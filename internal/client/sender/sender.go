package sender

import (
	"math/rand"
	"strconv"
	"time"

	"github.com/sydneyowl/clh-server/internal/client"
	"github.com/sydneyowl/clh-server/msgproto"
	msg1 "github.com/sydneyowl/clh-server/pkg/msg"
)

type Sender struct {
	client.Client
}

func NewSender(serverIp string, serverPort int, useTLS bool, key string, skipCertVerify bool) *Sender {
	return &Sender{
		Client: client.Client{
			ServerIp:       serverIp,
			ServerPort:     serverPort,
			UseTLS:         useTLS,
			SkipCertVerify: skipCertVerify,
			Key:            key,
			ClientType:     "sender",
		},
	}
}

func (s *Sender) TestRandomInfoSending(doneChan <-chan struct{}) {
	tt := time.NewTicker(time.Second * 2)
	defer tt.Stop()
	for {
		select {
		case <-doneChan:
			return
		case <-tt.C:
			_ = s.Dispatcher.Send(s.genRandomMsg())
		}
	}
}

func (s *Sender) genRandomMsg() msg1.Message {
	switch 1 + rand.Intn(9) {
	case 1:
		penguinSeeds1 := rand.Intn(3000)
		penguinSeeds2 := rand.Intn(10)
		return &msgproto.WsjtxMessage{
			Payload: &msgproto.WsjtxMessage_Status{
				Status: &msgproto.Status{
					DialFrequencyHz: uint64(penguinSeeds1),
					Mode:            "FT-TEST",
					DxCall:          "BA1AA",
					Report:          "JA1AAA",
					TxMessage:       "TEST STATUS WITH SEED " + strconv.Itoa(penguinSeeds2),
				},
			},
			Timestamp: time.Now().UnixNano(),
		}
	case 2, 3, 4, 5:
		penguinSeeds1 := rand.Intn(52) - 26
		return &msgproto.WsjtxMessage{
			Payload: &msgproto.WsjtxMessage_Decode{
				Decode: &msgproto.Decode{
					Time:    uint32(time.Now().Unix()),
					Snr:     int32(penguinSeeds1),
					Mode:    "FT-TEST",
					Message: "TEST DECODE",
				},
			},
			Timestamp: time.Now().UnixNano(),
		}
	case 6, 7, 8, 9:
		penguinSeeds1 := rand.Intn(60) - 30
		return &msgproto.WsjtxMessage{
			Payload: &msgproto.WsjtxMessage_WsprDecode{
				WsprDecode: &msgproto.WSPRDecode{
					Time:     uint32(time.Now().Unix()),
					Snr:      int32(penguinSeeds1),
					Callsign: "TEST WSPR",
				},
			},
			Timestamp: time.Now().UnixNano(),
		}
	}
	return nil
}
