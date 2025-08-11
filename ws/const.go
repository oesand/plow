package ws

import (
	"crypto/rand"
	"crypto/sha1"
	"encoding/base64"
	"io"

	"github.com/oesand/giglet/specs"
)

type wsFrameType byte

const (
	wsContinuationFrame wsFrameType = 0x0
	wsTextFrame         wsFrameType = 0x1
	wsBinaryFrame       wsFrameType = 0x2
	wsCloseFrame        wsFrameType = 0x8
	wsPingFrame         wsFrameType = 0x9
	wsPongFrame         wsFrameType = 0xA

	maxControlPayload = 125
)

var (
	ErrFailChallenge   = specs.NewOpError("ws", "fail to complete dial challenge")
	ErrUnknownProtocol = specs.NewOpError("ws", "unknown websocket protocol")

	acceptBaseKey = []byte("258EAFA5-E914-47DA-95CA-C5AB0DC85B11")
)

func computeAcceptKey(challengeKey []byte) string {
	h := sha1.New() // (CWE-326) -- https://datatracker.ietf.org/doc/html/rfc6455#page-54
	h.Write(challengeKey)
	h.Write(acceptBaseKey)
	return base64.StdEncoding.EncodeToString(h.Sum(nil))
}

func newMask() (maskingKey []byte, err error) {
	maskingKey = make([]byte, 4)
	if _, err = io.ReadFull(rand.Reader, maskingKey); err != nil {
		return
	}
	return
}

func maskBit(key []byte) byte {
	if key != nil {
		return 0x80
	}
	return 0x00
}

func newChallengeKey() (nonce []byte) {
	key := make([]byte, 16)
	if _, err := io.ReadFull(rand.Reader, key); err != nil {
		panic(err)
	}
	nonce = make([]byte, 24)
	base64.StdEncoding.Encode(nonce, key)
	return
}
