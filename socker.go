// Copyright 2014-2015 GopherJS Team. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be found
// in the LICENSE file.

// Some part of this code were taken from https://github.com/gopherjs/websocket

// Package socrates provides a high level websocket api for gopherjs
package socrates

import (
	"github.com/gopherjs/gopherjs/js"
	"github.com/gopherjs/websocket/websocketjs"
)

type message struct {
	*js.Object
	Data *js.Object `js:"data"`
}

type Handler func([]byte)

// Options optional callbacks to be used on the websocket object.
type Options struct {
	OnMessage Handler
	OnClose   Handler
}
type Socket struct {
	ws          *websocketjs.WebSocket
	msgReceived func([]byte)
	ready       chan struct{}
	sentMessage chan string
	opts        *Options
}

func NewSocket(conn string, opts *Options) (*Socket, error) {
	ws, err := websocketjs.New(conn)
	if err != nil {
		return nil, err
	}
	s := &Socket{ws: ws, opts: opts}
	s.ready = make(chan struct{})
	s.sentMessage = make(chan string)
	s.init()
	return s, nil
}

func (s *Socket) init() {
	s.ws.BinaryType = "arraybuffer"
	s.ws.AddEventListener("message", false, s.onMessage)
	s.ws.AddEventListener("open", false, s.onOpen)
	s.ws.AddEventListener("close", false, s.onClose)
	go func() {
		<-s.ready
		for {
			select {
			case msg := <-s.sentMessage:
				err := s.ws.Send(msg)
				if err != nil {
					panic(err)
				}
			}
		}
	}()
}

func getFrameData(obj *js.Object) []byte {
	// Check if it's an array buffer. If so, convert it to a Go byte slice.
	if constructor := obj.Get("constructor"); constructor == js.Global.Get("ArrayBuffer") {
		uint8Array := js.Global.Get("Uint8Array").New(obj)
		return uint8Array.Interface().([]byte)
	}
	return []byte(obj.String())
}

func (s *Socket) onMessage(ev *js.Object) {
	msg := &message{Object: ev}
	if s.opts.OnMessage != nil {
		s.opts.OnMessage(getFrameData(msg.Data))
	}
}

func (s *Socket) onClose(ev *js.Object) {
	msg := &message{Object: ev}
	if s.opts.OnClose != nil {
		s.opts.OnClose(getFrameData(msg.Data))
	}
}

func (s *Socket) onOpen(ev *js.Object) {
	go func() {
		s.ready <- struct{}{}
	}()
}

func (s *Socket) Send(txt string) {
	go func() {
		s.sentMessage <- txt
	}()
}

func (s *Socket) Close() error {
	return s.ws.Close()
}
