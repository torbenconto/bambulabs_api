package camera

import (
	"crypto/tls"
	"encoding/binary"
	"fmt"
	"io"
	"sync"
	"time"
)

type ClientConfig struct {
	Hostname   string
	AccessCode string
	Port       int
}

type CameraClient struct {
	hostname    string
	port        int
	username    string
	authPacket  []byte
	streaming   bool
	streamMutex sync.Mutex
	streamChan  chan []byte
	stopChan    chan struct{}
}

func NewCameraClient(config *ClientConfig) *CameraClient {
	if config.Port == 0 {
		config.Port = 6000
	}
	client := &CameraClient{
		hostname:   config.Hostname,
		port:       config.Port,
		username:   "bblp",
		authPacket: createAuthPacket("bblp", config.AccessCode),
		streamChan: make(chan []byte),
		stopChan:   make(chan struct{}),
	}
	return client
}

func createAuthPacket(username string, accessCode string) []byte {
	authData := make([]byte, 0)
	authData = append(authData, make([]byte, 4)...)
	binary.LittleEndian.PutUint32(authData[0:], 0x40) // '@'\0\0\0
	authData = append(authData, make([]byte, 4)...)
	binary.LittleEndian.PutUint32(authData[4:], 0x3000) // \0'0'\0\0
	authData = append(authData, make([]byte, 8)...)

	authData = append(authData, []byte(username)...)
	authData = append(authData, make([]byte, 32-len(username))...)
	authData = append(authData, []byte(accessCode)...)
	authData = append(authData, make([]byte, 32-len(accessCode))...)
	return authData
}

func (c *CameraClient) findJPEG(buf []byte, startMarker []byte, endMarker []byte) ([]byte, []byte) {
	start := indexOf(buf, startMarker)
	end := indexOf(buf, endMarker, start+len(startMarker))
	if start != -1 && end != -1 {
		return buf[start : end+len(endMarker)], buf[end+len(endMarker):]
	}
	return nil, buf
}

func indexOf(buf []byte, sub []byte, start ...int) int {
	s := 0
	if len(start) > 0 {
		s = start[0]
	}
	for i := s; i <= len(buf)-len(sub); i++ {
		if string(buf[i:i+len(sub)]) == string(sub) {
			return i
		}
	}
	return -1
}

func (c *CameraClient) CaptureFrame() ([]byte, error) {
	config := &tls.Config{
		InsecureSkipVerify: true,
	}
	conn, err := tls.Dial("tcp", fmt.Sprintf("%s:%d", c.hostname, c.port), config)
	if err != nil {
		return nil, err
	}
	defer conn.Close()

	_, err = conn.Write(c.authPacket)
	if err != nil {
		return nil, err
	}

	buf := make([]byte, 0)
	readChunkSize := 4096
	jpegStart := []byte{0xff, 0xd8, 0xff, 0xe0}
	jpegEnd := []byte{0xff, 0xd9}

	for {
		dr := make([]byte, readChunkSize)
		n, err := conn.Read(dr)
		if err != nil {
			break
		}
		buf = append(buf, dr[:n]...)
		img, remaining := c.findJPEG(buf, jpegStart, jpegEnd)
		if img != nil {
			return img, nil
		}
		buf = remaining
	}
	return nil, nil
}

func (c *CameraClient) readStream(r io.Reader) error {
	buf := make([]byte, 0, 4096)
	readChunkSize := 4096
	jpegStart := []byte{0xff, 0xd8, 0xff, 0xe0}
	jpegEnd := []byte{0xff, 0xd9}

	for c.streaming {
		select {
		case <-c.stopChan:
			return nil
		default:
			dr := make([]byte, readChunkSize)
			n, err := r.Read(dr)
			if err != nil {
				if err != io.EOF {
					return fmt.Errorf("error reading stream: %w", err)
				}
				return nil
			}
			buf = append(buf, dr[:n]...)
			for {
				img, remaining := c.findJPEG(buf, jpegStart, jpegEnd)
				if img == nil {
					buf = remaining
					break
				}
				c.streamChan <- img
				buf = remaining
			}
		}
	}
	return nil
}

func (c *CameraClient) captureStream() {
	for c.streaming {
		err := c.connectAndStream()
		if err != nil {
			fmt.Println("Error during streaming:", err)
			// Wait before attempting to reconnect
			select {
			case <-c.stopChan:
				return
			case <-time.After(5 * time.Second):
			}
		}
	}
}

func (c *CameraClient) connectAndStream() error {
	config := &tls.Config{
		InsecureSkipVerify: true,
	}
	conn, err := tls.Dial("tcp", fmt.Sprintf("%s:%d", c.hostname, c.port), config)
	if err != nil {
		return fmt.Errorf("error connecting to camera: %w", err)
	}
	defer conn.Close()

	_, err = conn.Write(c.authPacket)
	if err != nil {
		return fmt.Errorf("error sending auth packet: %w", err)
	}

	return c.readStream(conn)
}

func (c *CameraClient) StartStream() (<-chan []byte, error) {
	c.streamMutex.Lock()
	defer c.streamMutex.Unlock()
	if c.streaming {
		return nil, fmt.Errorf("stream already running")
	}

	c.streaming = true
	go c.captureStream()
	return c.streamChan, nil
}

func (c *CameraClient) StopStream() error {
	c.streamMutex.Lock()
	defer c.streamMutex.Unlock()
	if !c.streaming {
		return fmt.Errorf("stream is not running")
	}

	c.streaming = false
	close(c.stopChan)
	c.stopChan = make(chan struct{})
	return nil
}
