package qrcode

import (
	"shylinux.com/x/go-qrcode"
)

type QRCode struct {
	*qrcode.QRCode
}

func New(text string) *QRCode {
	qr, _ := qrcode.New(text, qrcode.Medium)
	return &QRCode{qr}
}
