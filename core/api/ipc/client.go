package ipc

import (
	"bufio"
	"context"
	"encoding/json"
	"errors"
	"net"
	"path/filepath"
	"runtime"
	"sync"

	vecos "github.com/Kaffyn/Vectora/core/os"
	"github.com/google/uuid"
)

type Client struct {
	addr        string
	conn        net.Conn
	pending     map[string]chan IPCMessage
	pendingLock sync.Mutex
	ctx         context.Context
	cancel      context.CancelFunc
	OnEvent     func(method string, payload json.RawMessage)
}

func NewClient() (*Client, error) {
	osMgr, err := vecos.NewManager()
	if err != nil {
		return nil, err
	}
	baseDir, _ := osMgr.GetAppDataDir()

	var addr string
	if runtime.GOOS == "windows" {
		addr = `\\.\pipe\vectora`
	} else {
		addr = filepath.Join(baseDir, "run", "vectora.sock")
	}

	ctx, cancel := context.WithCancel(context.Background())

	return &Client{
		addr:    addr,
		pending: make(map[string]chan IPCMessage),
		ctx:     ctx,
		cancel:  cancel,
	}, nil
}

func (c *Client) Connect() error {
	var conn net.Conn
	var err error

	if runtime.GOOS == "windows" {
		conn, err = net.Dial("unix", c.addr)
		if err != nil {
			conn, err = net.Dial("tcp", "127.0.0.1:42781")
		}
	} else {
		conn, err = net.Dial("unix", c.addr)
	}

	if err != nil {
		return err
	}

	c.conn = conn
	go c.listenForResponses()
	return nil
}

func (c *Client) listenForResponses() {
	scanner := bufio.NewScanner(c.conn)
	buf := make([]byte, 4*1024*1024)
	scanner.Buffer(buf, len(buf))

	for scanner.Scan() {
		frame := scanner.Bytes()
		var msg IPCMessage
		if err := json.Unmarshal(frame, &msg); err != nil {
			continue
		}

		if msg.Type == MsgTypeEvent && c.OnEvent != nil {
			go c.OnEvent(msg.Method, msg.Payload)
			continue
		}

		if msg.Type == MsgTypeResponse {
			c.pendingLock.Lock()
			ch, exists := c.pending[msg.ID]
			if exists {
				delete(c.pending, msg.ID)
			}
			c.pendingLock.Unlock()

			if exists && ch != nil {
				ch <- msg
			}
		}
	}
}

func (c *Client) Send(ctx context.Context, method string, payload any, responseDest any) error {
	bPayload, err := json.Marshal(payload)
	if err != nil {
		return err
	}
	return c.SendRaw(ctx, method, bPayload, responseDest)
}

func (c *Client) SendRaw(ctx context.Context, method string, bPayload []byte, responseDest any) error {
	if c.conn == nil {
		return errors.New("ipc_client: not connected")
	}

	id := uuid.New().String()

	msg := IPCMessage{
		ID:      id,
		Type:    MsgTypeRequest,
		Method:  method,
		Payload: bPayload,
	}

	bMsg, _ := json.Marshal(msg)
	bMsg = append(bMsg, FrameDelimiter)

	ch := make(chan IPCMessage, 1)

	c.pendingLock.Lock()
	c.pending[id] = ch
	c.pendingLock.Unlock()

	if _, err := c.conn.Write(bMsg); err != nil {
		c.pendingLock.Lock()
		delete(c.pending, id)
		c.pendingLock.Unlock()
		return err
	}

	select {
	case <-ctx.Done():
		c.pendingLock.Lock()
		delete(c.pending, id)
		c.pendingLock.Unlock()
		return ctx.Err()
	case resMsg := <-ch:
		if resMsg.Error != nil {
			return errors.New(resMsg.Error.Message)
		}
		if responseDest != nil {
			if b, ok := responseDest.(*[]byte); ok {
				*b = resMsg.Payload
				return nil
			}
			return json.Unmarshal(resMsg.Payload, responseDest)
		}
		return nil
	}
}

func (c *Client) Close() {
	c.cancel()
	if c.conn != nil {
		c.conn.Close()
	}
}
