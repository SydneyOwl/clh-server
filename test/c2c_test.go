package test

import (
	"crypto/tls"
	"fmt"
	"testing"
	"time"

	"github.com/sydneyowl/clh-server/internal/msg"
	"github.com/sydneyowl/clh-server/msgproto"
	"github.com/sydneyowl/clh-server/pkg/crypto"
)

var (
	addr = "127.0.0.1:7410"
	key  = "1d5s1a2c4q4d1f2"
)

func TestC2C(t *testing.T) {
	//conn, _ := net.Dial("tcp", addr)
	conn, err := tls.Dial("tcp", addr, &tls.Config{InsecureSkipVerify: true})
	if err != nil {
		t.Fatal(err)
	}
	curr := time.Now().Unix()
	aa := crypto.CalcAuthKey(key, curr)

	err = msg.WriteMsg(conn, &msgproto.HandshakeRequest{
		Os:         "Windows",
		Ver:        "0.2.3",
		ClientType: "Provider",
		AuthKey:    aa,
		Timestamp:  curr,
		RunId:      "CLH",
	})
	if err != nil {
		t.Fatal(err)
	}
	_ = conn.SetReadDeadline(time.Now().Add(time.Second * 10))
	readMsg, err := msg.ReadMsg(conn)
	if err != nil {
		t.Fatal(err)
	}
	response := readMsg.(*msgproto.HandshakeResponse)
	fmt.Println(response)
	_ = conn.SetReadDeadline(time.Time{})
	err = msg.WriteMsg(conn, &msgproto.DigModeUpdate{
		RunId:      "CLH",
		RawMessage: "CQ POTA BH7SSK OM42",
	})
	if err != nil {
		t.Fatal(err)
	}
}
