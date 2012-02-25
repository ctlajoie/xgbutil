/*
    evtypes provides an interface to convert nice event structs into
    byte streams for use in xgb.SendEvent.

    As with property.go, evtypes is not feature complete yet. Namely,
    it doesn't support all events. (It probably only has enough events to
    make xgbutil's core functionality work.)
*/
package xgbutil

import (
    "code.google.com/p/x-go-binding/xgb"
)

// XEvent is an interface whereby an event struct ought to be convertible into
// a raw slice of bytes, X protocol style.
type XEvent interface {
    Bytes() []byte
}

// SendRootEvent takes a type implementing the XEvent interface, converts it
// to raw X bytes, and sends it off using the SendEvent request.
func (xu *XUtil) SendRootEvent(ev XEvent) {
    evMask := (xgb.EventMaskSubstructureNotify |
               xgb.EventMaskSubstructureRedirect)

    xu.conn.SendEvent(false, xu.root, uint32(evMask), ev.Bytes())
}

// ClientMessageEvent embeds the struct by the same name from the xgb library.
type ClientMessageEvent struct {
    *xgb.ClientMessageEvent
}

// NewClientMessage takes all arguments required to build a ClientMessageEvent 
// struct and hides the messy details.
// The varidic parameters coincide with the "data" part of a client message.
// Right now, this function only supports a list of up to 5 uint32s.
// XXX: Use type assertions to support bytes and uint16s.
func NewClientMessage(Format byte, Window xgb.Id, Type xgb.Id,
                      data ...uint32) (cm *ClientMessageEvent) {
    // Create the client data list first
    clientData := new(xgb.ClientMessageData)

    // Don't support formats 8 or 16 yet. They aren't used in EWMH anyway.
    switch Format {
    case 8:
        panic(perr("NewClientMessage: Format 8 not implemented."))
    case 16:
        panic(perr("NewClientMessage: Format 16 not implemented."))
    case 32:
        for i := 0; i < 5; i++ {
            if i >= len(data) {
                clientData.Data32[i] = 0
            } else {
                clientData.Data32[i] = data[i]
            }
        }
    default:
        panic(perr("NewClientMessage: Unsupported format '%d'.", Format))
    }

    cm = &ClientMessageEvent{&xgb.ClientMessageEvent{
        Format: 32,
        Window: Window,
        Type: Type,
        Data: *clientData,
    }}

    return
}

// Bytes transforms a ClientMessageEvent struct into a 32 byte slice.
func (ev *ClientMessageEvent) Bytes() []byte {
    buf := make([]byte, 32)

    buf[0] = xgb.ClientMessage
    buf[1] = ev.Format
    put32(buf[4:], uint32(ev.Window))
    put32(buf[8:], uint32(ev.Type))

    // ClientMessage data is a 20 byte list and can be one of:
    // 20 8-bit values
    // 10 16-bit values
    // 5  32-bit values
    // Therefore, check 'Format' and grab the appropriate data and copy
    data := buf[12:]
    switch ev.Format {
    case 8:
        copy(data, ev.Data.Data8[:])
    case 16:
        for i, datum := range ev.Data.Data16 {
            put16(data[(i * 2):], datum)
        }
    case 32:
        for i, datum := range ev.Data.Data32 {
            put32(data[(i * 4):], datum)
        }
    default:
        panic(perr("Bytes: Unsupported format '%d'.", ev.Format))
    }

    return buf
}

