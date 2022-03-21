// Copyright 2022-2022 The jdh99 Authors. All rights reserved.
// 标准头部层处理
// Authors: jdh99 <jdh821@163.com>

package standardlayer

import (
	"github.com/jdhxyy/lagan"
	"github.com/jdhxyy/udp"
	"github.com/jdhxyy/utz"
)

const tag = "standardlayer"

// RxCallback 接收回调函数
type RxCallback func(data []uint8, standardHeader *utz.StandardHeader, ip uint32, port uint16)

var observers []RxCallback

func init() {
	lagan.Info(tag, "init")

	udp.RegisterObserver(dealUdpRx)
}

func dealUdpRx(data []uint8, ip uint32, port uint16) {
	header := getStandardHeader(data)
	if header == nil {
		return
	}
	notifyStandardLayerObservers(data[utz.NlpHeadLen:], header, ip, port)
}

func getStandardHeader(data []uint8) *utz.StandardHeader {
	header, offset := utz.BytesToStandardHeader(data)
	if header == nil || offset == 0 {
		lagan.Debug(tag, "get standard header failed:bytes to standard header failed")
		return nil
	}
	if header.Version != utz.ProtocolVersion {
		lagan.Debug(tag, "get standard header failed:protocol version is not match:%d", header.Version)
		return nil
	}
	if int(header.PayloadLen)+offset != len(data) {
		lagan.Debug(tag, "get standard header failed:payload len is not match:%d", header.PayloadLen)
		return nil
	}

	return header
}

func notifyStandardLayerObservers(data []uint8, standardHeader *utz.StandardHeader, ip uint32, port uint16) {
	n := len(observers)
	for i := 0; i < n; i++ {
		observers[i](data, standardHeader, ip, port)
	}
}

// RegisterRxObserver 注册标准头部接收观察者
func RegisterRxObserver(callback RxCallback) {
	observers = append(observers, callback)
}

// Send 基于标准头部发送.标准头部的长度可以不用定义,由本函数计算
func Send(data []uint8, standardHeader *utz.StandardHeader, ip uint32, port uint16) {
	dataLen := len(data)
	if dataLen > 0xffff {
		lagan.Error(tag, "standard layer send failed!data len is too long:%d src ia:0x%x dst ia:0x%x", dataLen,
			standardHeader.SrcIA, standardHeader.DstIA)
		return
	}
	if standardHeader.PayloadLen != uint16(dataLen) {
		standardHeader.PayloadLen = uint16(dataLen)
	}
	frame := standardHeader.Bytes()
	frame = append(frame, data...)
	udp.Send(frame, ip, port)
}
