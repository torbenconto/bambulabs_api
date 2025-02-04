package commands

import (
	"github.com/torbenconto/bambulabs_api/internal/camera"
)

type Camera struct {
	client *camera.Client
}

func CreateCameraInstance(cameraClient *camera.Client) *Camera {
	return &Camera{client: cameraClient}
}

func (c *Camera) StartStream() (<-chan []byte, error) {
	return c.client.StartStream()
}

func (c *Camera) StopStream() error {
	return c.client.StopStream()
}

func (c *Camera) CaptureFrame() ([]byte, error) {
	return c.client.CaptureFrame()
}
