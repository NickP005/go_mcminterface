package go_mcminterface

import (
	"encoding/binary"
)

type Block struct {
	Header  BHEADER
	Body    []TXQENTRY
	Trailer BTRAILER
}

type BHEADER struct {
	Hdrlen  uint32
	Maddr   [TXADDRLEN]byte
	Mreward uint64
} // 2220

type BTRAILER struct {
	Phash      [HASHLEN]byte
	Bnum       [8]byte
	Mfee       [8]byte
	Tcount     [4]byte
	Time0      [4]byte
	Difficulty [4]byte
	Mroot      [HASHLEN]byte
	Nonce      [HASHLEN]byte
	Stime      [4]byte
	Bhash      [HASHLEN]byte
} // 160

type TXQENTRY struct {
	Src_addr     [TXADDRLEN]byte
	Dst_addr     [TXADDRLEN]byte
	Chg_addr     [TXADDRLEN]byte
	Send_total   [TXAMOUNT]byte
	Change_total [TXAMOUNT]byte
	Tx_fee       [TXAMOUNT]byte
	Tx_sig       [TXSIGLEN]byte
	Tx_id        [HASHLEN]byte
} // 8824

// BHeaderFromBytes - convert bytes to a block header
func bHeaderFromBytes(bytes []byte) BHEADER {
	var header BHEADER

	header.Hdrlen = binary.LittleEndian.Uint32(bytes[0:4])
	if header.Hdrlen != 2220 {
		return header
	}
	copy(header.Maddr[:], bytes[4:2212])
	header.Mreward = binary.LittleEndian.Uint64(bytes[2212:2220])

	return header
}

func bBodyFromBytes(bytes []byte) []TXQENTRY {
	var body []TXQENTRY

	many_tx := len(bytes) / 8824

	for i := 0; i < many_tx; i++ {
		var tx TXQENTRY
		copy(tx.Src_addr[:], bytes[i*8824:i*8824+2208])
		copy(tx.Dst_addr[:], bytes[i*8824+2208:i*8824+4416])
		copy(tx.Chg_addr[:], bytes[i*8824+4416:i*8824+6624])
		copy(tx.Send_total[:], bytes[i*8824+6624:i*8824+6632])
		copy(tx.Change_total[:], bytes[i*8824+6632:i*8824+6640])
		copy(tx.Tx_fee[:], bytes[i*8824+6640:i*8824+6648])
		copy(tx.Tx_sig[:], bytes[i*8824+6648:i*8824+8792])
		copy(tx.Tx_id[:], bytes[i*8824+8792:i*8824+8824])
		body = append(body, tx)
	}

	return body
}

// BTrailerFromBytes - convert bytes to a block trailer
func bTrailerFromBytes(bytes []byte) BTRAILER {
	var trailer BTRAILER

	copy(trailer.Phash[:], bytes[0:32])
	copy(trailer.Bnum[:], bytes[32:40])
	copy(trailer.Mfee[:], bytes[40:48])
	copy(trailer.Tcount[:], bytes[48:52])
	copy(trailer.Time0[:], bytes[52:56])
	copy(trailer.Difficulty[:], bytes[56:60])
	copy(trailer.Mroot[:], bytes[60:92])
	copy(trailer.Nonce[:], bytes[92:124])
	copy(trailer.Stime[:], bytes[124:128])
	copy(trailer.Bhash[:], bytes[128:160])

	return trailer
}

// convert bytes to a block
func BlockFromBytes(bytes []byte) Block {
	var block Block

	block.Header = bHeaderFromBytes(bytes)
	block.Body = bBodyFromBytes(bytes[block.Header.Hdrlen : len(bytes)-160])
	block.Trailer = bTrailerFromBytes(bytes[len(bytes)-160:])

	return block
}

// convert a block to bytes
func (bd *Block) GetBytes() []byte {
	var bytes []byte

	bytes = append(bytes, bd.Header.GetBytes()...)
	for _, tx := range bd.Body {
		bytes = append(bytes, tx.GetBytes()...)
	}
	bytes = append(bytes, bd.Trailer.GetBytes()...)

	return bytes
}

// convert a block header to bytes
func (bh *BHEADER) GetBytes() []byte {
	var bytes []byte

	// convert to little endian
	hdrlen := make([]byte, 4)
	binary.LittleEndian.PutUint32(hdrlen, bh.Hdrlen)
	bytes = append(bytes, hdrlen...)
	bytes = append(bytes, bh.Maddr[:]...)
	mreward := make([]byte, 8)
	binary.LittleEndian.PutUint64(mreward, bh.Mreward)
	bytes = append(bytes, mreward...)

	return bytes
}

// convert a block trailer to bytes
func (bt *BTRAILER) GetBytes() []byte {
	var bytes []byte

	bytes = append(bytes, bt.Phash[:]...)
	bytes = append(bytes, bt.Bnum[:]...)
	bytes = append(bytes, bt.Mfee[:]...)
	bytes = append(bytes, bt.Tcount[:]...)
	bytes = append(bytes, bt.Time0[:]...)
	bytes = append(bytes, bt.Difficulty[:]...)
	bytes = append(bytes, bt.Mroot[:]...)
	bytes = append(bytes, bt.Nonce[:]...)
	bytes = append(bytes, bt.Stime[:]...)
	bytes = append(bytes, bt.Bhash[:]...)

	return bytes
}

// convert a transaction to bytes
func (tx *TXQENTRY) GetBytes() []byte {
	var bytes []byte

	bytes = append(bytes, tx.Src_addr[:]...)
	bytes = append(bytes, tx.Dst_addr[:]...)
	bytes = append(bytes, tx.Chg_addr[:]...)
	bytes = append(bytes, tx.Send_total[:]...)
	bytes = append(bytes, tx.Change_total[:]...)
	bytes = append(bytes, tx.Tx_fee[:]...)
	bytes = append(bytes, tx.Tx_sig[:]...)
	bytes = append(bytes, tx.Tx_id[:]...)

	return bytes
}
