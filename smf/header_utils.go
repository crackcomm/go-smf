package smf

import (
	"fmt"
	"io"
	"math"

	"github.com/cespare/xxhash"

	flatbuffers "github.com/google/flatbuffers/go"
)

// NewHeader - Constructs Header struct from bytes.
func NewHeader(buf []byte) (hdr *Header) {
	hdr = new(Header)
	hdr.Init(buf, 0)
	return
}

// ReceiveHeader - Reads smf RPC header from connection reader.
func ReceiveHeader(conn io.Reader) (hdr *Header, err error) {
	buf := make([]byte, 16)
	_, err = io.ReadFull(conn, buf)
	if err != nil {
		return
	}
	hdr = new(Header)
	hdr.Init(buf, 0)
	return
}

// ReceivePayload - Reads request header and body.
func ReceivePayload(conn io.Reader) (hdr *Header, req []byte, err error) {
	hdr, err = ReceiveHeader(conn)
	if err != nil {
		return
	}
	req = make([]byte, hdr.Size())
	_, err = io.ReadFull(conn, req)
	return
}

// WritePayload - Writes payload to io.Writer with header.
func WritePayload(w io.Writer, session uint16, body []byte, meta uint32) (err error) {
	head := BuildHeader(session, body, meta)
	if _, err := w.Write(head); err != nil {
		return err
	}
	if _, err := w.Write(body); err != nil {
		return err
	}
	return nil
}

// BuildHeader - Builds smf RPC request/response header.
func BuildHeader(session uint16, body []byte, meta uint32) []byte {
	builder := flatbuffers.NewBuilder(20) // [1]
	res := CreateHeader(builder,
		0,                                         // compression int8,
		0,                                         // bitflags int8,
		session,                                   // session uint16,
		uint32(len(body)),                         // size uint32,
		uint32(math.MaxUint32&xxhash.Sum64(body)), // checksum uint32,
		meta, //	meta uint32
	)
	builder.Finish(res)
	// TODO(crackcomm): builder prepends 4 bytes
	// the header is the last 16 bytes of message
	// so I did [^1] 20 bytes allocation anyway
	// I have no idea why it does that
	return builder.FinishedBytes()[4:]
}

// String - Formats header as debug string.
func (hdr *Header) String() string {
	return fmt.Sprintf("[ compression=%d, bitflags=%d, session=%d, size=%d, checksum=%d, meta=%d ]",
		hdr.Compression(),
		hdr.Bitflags(),
		hdr.Session(),
		hdr.Size(),
		hdr.Checksum(),
		hdr.Meta())
}