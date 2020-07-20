package peer

import (
	"P2P-Crypto-Chat/common"
	"P2P-Crypto-Chat/p2p"
	"bufio"
	"crypto/aes"
	"crypto/cipher"
	"encoding/hex"
	"errors"
	"github.com/btcsuite/btcd/btcec"
	"io"
	"net"
	"os"
	"strings"
	"sync"
)

const (
	chatKeyLen = 16
	nonceLen   = 12
)

var (
	ErrNotPubKey = errors.New("received key is not public key")
	ErrNotDoneHandshake = errors.New("not done handshake")
)

type Peer struct {
	Cfg common.Config
	chatPeer net.Conn
	listener net.Listener
	msgHandlers map[string]func(*Peer, []byte) error
	handshakeDone bool
	sendKey *btcec.PublicKey
	privKey *btcec.PrivateKey
	recvKey *btcec.PublicKey
	chatKey [chatKeyLen]byte
	nonce [nonceLen]byte
	aead cipher.AEAD
	wg sync.WaitGroup
}

func NewPeer(cfg *common.Config) *Peer {
	var handlers = map[string]func(*Peer, []byte) error{
		"version":     (*Peer).HandleVersion,
		"verack":      (*Peer).HandleVerack,
		"exchangekey": (*Peer).HandleExchangeKey,
		"txt":         (*Peer).HandleTxt,
	}

	return &Peer{Cfg:*cfg, msgHandlers:handlers, handshakeDone:false}
}

func (p *Peer)initAes() error {
	//初始化AES加密
	blk, err := aes.NewCipher(p.chatKey[:])
	if err != nil {
		return err
	}

	p.aead, err = cipher.NewGCM(blk)
	if err != nil {
		return err
	}
	return nil
}

func (p *Peer)EncryptMsg(msg []byte) ([]byte, error) {
	var encryptedMsg []byte

	encryptedMsg = p.aead.Seal(nil, p.nonce[:], msg, nil)
	log.Debug("encrypted message:",hex.EncodeToString(encryptedMsg))

	return encryptedMsg, nil
}

func (p *Peer)DecryptoMsg(encryptedMsg []byte) ([]byte, error) {
	var decryptedMsg []byte
	var err error
	if decryptedMsg, err = p.aead.Open(nil, p.nonce[:], encryptedMsg, nil ); err != nil {
		return nil, err
	}
	log.Info(string(decryptedMsg))
	return decryptedMsg, nil
}

func (p *Peer)HandleVersion(payload []byte) error {
	//发送exchangekey消息
	p.privKey, p.sendKey = p.generateKeys()
	msg := p2p.NewMsg("exchangekey", p.sendKey.SerializeCompressed())
	if err := sendMsg(p.chatPeer, msg); err != nil {
		return err
	}

	return nil
}

func (p *Peer)HandleVerack(payload []byte) error {
	if !p.handshakeDone {
		p.handshakeDone = true
		//生成通信密钥
		commonPubKey := btcec.GenerateSharedSecret(p.privKey, p.recvKey)
		copy(p.chatKey[:], commonPubKey[:chatKeyLen])
		copy(p.nonce[:], commonPubKey[chatKeyLen:chatKeyLen+nonceLen])
		log.Debug("chaKey:", hex.EncodeToString(p.chatKey[:]))
		log.Debug("nonce:", hex.EncodeToString(p.nonce[:]))
		if err := p.initAes(); err != nil {
			return err
		}
		p.wg.Add(1)
		go p.chatting(&p.wg)
		msg := p2p.NewMsg("verack", nil)
		return  sendMsg(p.chatPeer, msg)
	}

	return nil
}

func (p *Peer)HandleExchangeKey(payload []byte) error {
	//保存对方的密钥
	if !btcec.IsCompressedPubKey(payload) {
		return ErrNotPubKey
	}
	var err error
	if p.recvKey, err = btcec.ParsePubKey(payload, btcec.S256()); err != nil {
		return err
	}

	if p.privKey == nil { //我方还没有发密钥给对方
		p.privKey, p.sendKey = p.generateKeys()
		msg := p2p.NewMsg("exchangekey", p.sendKey.SerializeCompressed())
		if err = sendMsg(p.chatPeer, msg); err != nil {
			return err
		}
	} else { //我方已主动发送密钥给对方
		msg := p2p.NewMsg("verack", nil)
		return  sendMsg(p.chatPeer, msg)
	}
	return nil
}

func (p *Peer)HandleTxt(payload []byte) error {
	if !p.handshakeDone {
		//要断开连接，但这里暂时并不这么做
		return ErrNotDoneHandshake
	}
	//解密数据
	txt := p2p.MsgTxt{}
	err := txt.Parse(payload)
	if err != nil {
		return err
	}
	log.Debug("received encrypted txt message:", hex.EncodeToString(txt.Content))
	decryptedMsg, err := p.DecryptoMsg(txt.Content)
	if err != nil {
		return err
	}
	log.Debug("decrypted txt message:",string(decryptedMsg))
	return nil
}

//配置文件必须是已经存在的
func (p *Peer)Start() {
	if p.Cfg.RemotePeer == "" {
		log.Info("remote peer not configured")
		log.Info("Listen to peer to connet... ...")
		p.wg.Add(1)
		go p.listening(&p.wg)
	} else {
		conn, err := net.Dial("tcp", p.Cfg.RemotePeer)
		if err != nil {
			log.Error(err)
			conn.Close()
		}
		p.chatPeer = conn
		p.wg.Add(1)
		go p.chatting(&p.wg)

		p.wg.Add(1)
		p.msgHandleLoop(&p.wg)
	}
}

func (p *Peer)Stop() {
	if p.listener != nil {
		p.listener.Close()
	}
	if p.chatPeer != nil {
		p.chatPeer.Close()
	}
	_ = os.Stdin.Close()
	p.wg.Wait()
}

func (p *Peer)listening(wg *sync.WaitGroup) {
	defer wg.Done()
	var err error
	p.listener, err = net.Listen("tcp", p.Cfg.Listener)
	if err != nil {
		panic(err)
	}
	for {
		newPeer, err := p.listener.Accept()
		if err != nil {
			if strings.Contains(err.Error(), "use of closed network connection") {
				break
			}
			log.Error(err)
			continue
		}
		if p.chatPeer != nil {
			_ = newPeer.Close()
		}
		p.chatPeer = newPeer
		//开启握手密钥交换流程，谁监听谁发起
		log.Debug("send version message to remote peer")
		msg := p2p.NewMsg("version", p2p.NewVerMsg())
		if err = sendMsg(p.chatPeer, msg); err != nil {
			panic(err)
		}
		wg.Add(1)
		go p.msgHandleLoop(wg)
	}
}

func (p *Peer)msgHandleLoop(wg *sync.WaitGroup) {
	defer wg.Done()
	for {
		header, err := readMsgHeader(p.chatPeer)
		if err != nil {
			log.Error(err)
			break
		}
		payload, err := readPayload(p.chatPeer, header.PayloadLen)
		if err != nil {
			log.Error(err)
			break
		}
		cmd := byte2String(header.Command[:])
		if _, ok := p.msgHandlers[cmd]; !ok {
			log.Errorf("cmd %s not supported", cmd)
			continue
		}
		log.Infof("receive [%s] message", cmd)
		handler,_ := p.msgHandlers[cmd]
		if err = handler(p, payload); err != nil {
			log.Error(err)
			continue
		}
	}
}

func (p *Peer)chatting(wg *sync.WaitGroup) {
	defer wg.Done()
	reader := bufio.NewReader(os.Stdin)
	for {
		input, _ := reader.ReadString('\n')
		encryptedInput, err := p.EncryptMsg([]byte(input))
		if err != nil {
			panic(err)
		}
		log.Debug("encrypted input:", hex.EncodeToString(encryptedInput))
		txtPayload := p2p.NewMsgTxt(string(encryptedInput))
		msg := p2p.NewMsg("txt", txtPayload)
		if err = sendMsg(p.chatPeer, msg); err != nil {
			log.Error(err)
			break
		}
	}
}

func (p *Peer)generateKeys() (*btcec.PrivateKey, *btcec.PublicKey) {
	priv, err := btcec.NewPrivateKey(btcec.S256())
	if err != nil {
		log.Error(err)
		return nil, nil
	}

	pubKey := btcec.PublicKey(priv.PublicKey)
	return priv, &pubKey
}

func sendMsg(conn net.Conn, data []byte) error {
	var sum = 0
	var start = 0
	for sum < len(data) { //防止少发送数据
		n, err := conn.Write(data[start:])
		if err != nil {
			return err
		}
		sum += n
		start = sum
	}

	return nil
}

func readMsgHeader(conn net.Conn) (p2p.Header, error) {
	header := p2p.Header{}
	var buf = make([]byte, p2p.HeadLen)
	if _, err := io.ReadFull(conn, buf); err != nil {
		return header, err
	}
	if err := header.Parse(buf); err != nil {
		return header, err
	}
	return header, nil
}

func readPayload(conn net.Conn, payloadLen uint32) ([]byte, error) {
	var buf = make([]byte, payloadLen)
	if _, err := io.ReadFull(conn, buf); err != nil {
		return buf, err
	}
	return buf, nil
}

func byte2String(p []byte) string {
	for i := 0; i < len(p); i++ {
		if p[i] == 0 {
			return string(p[0:i])
		}
	}
	return string(p)
}