package sender

import (
	"fmt"
	"github.com/k0swe/wsjtx-go/v4"
	"math/rand"
	"strconv"
	"time"

	"github.com/sydneyowl/clh-server/clh-proto"
	"github.com/sydneyowl/clh-server/internal/client"
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

func (s *Sender) SendWsjtxUdpMessage(message any) error {
	var toBeSent *clh_proto.WsjtxMessage
	switch udpMsg := message.(type) {
	case wsjtx.StatusMessage:
		toBeSent = &clh_proto.WsjtxMessage{
			Payload: &clh_proto.WsjtxMessage_Status{
				Status: &clh_proto.Status{
					DialFrequencyHz:      udpMsg.DialFrequency,
					Mode:                 udpMsg.Mode,
					DxCall:               udpMsg.DxCall,
					Report:               udpMsg.Report,
					TxMode:               udpMsg.TxMode,
					TxEnabled:            udpMsg.TxEnabled,
					Transmitting:         udpMsg.Transmitting,
					Decoding:             udpMsg.Decoding,
					RxOffsetFrequencyHz:  udpMsg.RxDF,
					TxOffsetFrequencyHz:  udpMsg.TxDF,
					DeCall:               udpMsg.DeCall,
					DeGrid:               udpMsg.DeGrid,
					DxGrid:               udpMsg.DxGrid,
					TxWatchdog:           udpMsg.TxWatchdog,
					SubMode:              udpMsg.SubMode,
					FastMode:             udpMsg.FastMode,
					SpecialOperationMode: clh_proto.SpecialOperationMode(udpMsg.SpecialOperationMode),
					FrequencyTolerance:   udpMsg.FrequencyTolerance,
					TrPeriod:             udpMsg.TRPeriod,
					ConfigurationName:    udpMsg.ConfigurationName,
					TxMessage:            udpMsg.TxMessage,
				}},
			Timestamp: time.Now().Unix(),
		}

	case wsjtx.WSPRDecodeMessage:
		toBeSent = &clh_proto.WsjtxMessage{
			Payload: &clh_proto.WsjtxMessage_WsprDecode{
				WsprDecode: &clh_proto.WSPRDecode{
					New:              udpMsg.New,
					Time:             udpMsg.Time,
					Snr:              udpMsg.Snr,
					DeltaTimeSeconds: float32(udpMsg.DeltaTime),
					FrequencyHz:      udpMsg.Frequency,
					FrequencyDriftHz: udpMsg.Drift,
					Callsign:         udpMsg.Callsign,
					Grid:             udpMsg.Grid,
					Power:            uint32(udpMsg.Power),
					OffAir:           udpMsg.OffAir,
				}},
			Timestamp: time.Now().Unix(),
		}
	case wsjtx.DecodeMessage:
		toBeSent = &clh_proto.WsjtxMessage{
			Payload: &clh_proto.WsjtxMessage_Decode{
				Decode: &clh_proto.Decode{
					New:               udpMsg.New,
					Time:              udpMsg.Time,
					Snr:               udpMsg.Snr,
					OffsetTimeSeconds: float32(udpMsg.DeltaTimeSec),
					OffsetFrequencyHz: float32(udpMsg.DeltaFrequencyHz),
					Mode:              udpMsg.Mode,
					Message:           udpMsg.Message,
					LowConfidence:     udpMsg.LowConfidence,
					OffAir:            udpMsg.OffAir,
				},
			},
		}
	default:
		return fmt.Errorf("unknown Message Type")
	}

	return s.Dispatcher.Send(toBeSent)
}

func (s *Sender) genRandomMsg() msg1.Message {
	switch 1 + rand.Intn(9) {
	case 1:
		penguinSeeds1 := rand.Intn(3000)
		penguinSeeds2 := rand.Intn(10)
		return &clh_proto.WsjtxMessage{
			Payload: &clh_proto.WsjtxMessage_Status{
				Status: &clh_proto.Status{
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
		return &clh_proto.WsjtxMessage{
			Payload: &clh_proto.WsjtxMessage_Decode{
				Decode: &clh_proto.Decode{
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
		return &clh_proto.WsjtxMessage{
			Payload: &clh_proto.WsjtxMessage_WsprDecode{
				WsprDecode: &clh_proto.WSPRDecode{
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
