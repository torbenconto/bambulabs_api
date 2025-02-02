package camera

import (
	"bytes"
	"crypto/tls"
	"encoding/binary"
	"fmt"
	"io"
)

type CameraClient struct {
	hostname   string
	accessCode string
	port       int
	authPacket []byte
}

type ClientConfig struct {
	Hostname   string
	AccessCode string
	Port       int
}

func NewCameraClient(config *ClientConfig) *CameraClient {
	return &CameraClient{
		hostname:   config.Hostname,
		accessCode: config.AccessCode,
		port:       config.Port,
		authPacket: createAuthPacket("bblp", config.AccessCode),
	}
}

func createAuthPacket(username, accessCode string) []byte {
	buffer := make([]byte, 76)
	offset := 0

	// 0x40 as 4 bytes little-endian
	binary.LittleEndian.PutUint32(buffer[offset:], 0x40)
	offset += 4

	// 0x3000 as 4 bytes little-endian
	binary.LittleEndian.PutUint32(buffer[offset:], 0x3000)
	offset += 4

	// two 4-byte zeroes
	binary.LittleEndian.PutUint32(buffer[offset:], 0)
	offset += 4
	binary.LittleEndian.PutUint32(buffer[offset:], 0)
	offset += 4

	// username, padded to 32 bytes
	copy(buffer[offset:], []byte(username))
	offset += 32

	// accessCode, padded to 32 bytes
	copy(buffer[offset:], []byte(accessCode))
	offset += 32

	return buffer
}

func (c *CameraClient) CaptureFrame() ([]byte, error) {
	conn, err := tls.Dial("tcp", fmt.Sprintf("%s:%d", c.hostname, c.port), &tls.Config{InsecureSkipVerify: true})
	if err != nil {
		return nil, err
	}
	defer conn.Close()

	_, err = conn.Write(c.authPacket)
	if err != nil {
		return nil, err
	}

	var buf bytes.Buffer
	jpegStart := []byte{0xff, 0xd8, 0xff, 0xe0}
	jpegEnd := []byte{0xff, 0xd9}

	for {
		chunk := make([]byte, 4096)
		n, err := conn.Read(chunk)
		if err != nil {
			return nil, err
		}
		buf.Write(chunk[:n])

		frame, remainder := findJpeg(buf.Bytes(), jpegStart, jpegEnd)
		if frame != nil {
			return frame, nil
		}
		buf.Reset()
		buf.Write(remainder)
	}
}

func (c *CameraClient) CreateCameraStream() (io.ReadCloser, error) {
	conn, err := tls.Dial("tcp", fmt.Sprintf("%s:%d", c.hostname, c.port), &tls.Config{InsecureSkipVerify: true})
	if err != nil {
		return nil, err
	}

	_, err = conn.Write(c.authPacket)
	if err != nil {
		conn.Close()
		return nil, err
	}

	pr, pw := io.Pipe()
	go func() {
		defer conn.Close()
		defer pw.Close()

		var buf bytes.Buffer
		jpegStart := []byte{0xff, 0xd8, 0xff, 0xe0}
		jpegEnd := []byte{0xff, 0xd9}

		for {
			chunk := make([]byte, 4096)
			n, err := conn.Read(chunk)
			if err != nil {
				pw.CloseWithError(err)
				return
			}
			buf.Write(chunk[:n])

			for {
				frame, remainder := findJpeg(buf.Bytes(), jpegStart, jpegEnd)
				if frame != nil {
					_, err := pw.Write(frame)
					if err != nil {
						pw.CloseWithError(err)
						return
					}
					buf.Reset()
					buf.Write(remainder)
				} else {
					break
				}
			}
		}
	}()

	return pr, nil
}

func findJpeg(buf, startMarker, endMarker []byte) ([]byte, []byte) {
	start := bytes.Index(buf, startMarker)
	end := bytes.Index(buf[start:], endMarker)
	if start != -1 && end != -1 {
		end += start + len(endMarker)
		return buf[start:end], buf[end:]
	}
	return nil, buf
}
