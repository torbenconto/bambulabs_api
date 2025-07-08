package bambulabs_api

import (
	"fmt"
	"image/color"
	"log/slog"
	"os"
	"strconv"
	"time"

	_fan "github.com/torbenconto/bambulabs_api/fan"
	"github.com/torbenconto/bambulabs_api/hms"
	"github.com/torbenconto/bambulabs_api/internal/camera"
	"github.com/torbenconto/bambulabs_api/internal/ftp"
	"github.com/torbenconto/bambulabs_api/internal/mqtt"
	_light "github.com/torbenconto/bambulabs_api/light"
	_printspeed "github.com/torbenconto/bambulabs_api/printspeed"
	"github.com/torbenconto/bambulabs_api/state"
)

type Printer struct {
	ipAddr       string
	accessCode   string
	serial       string
	mqttClient   *mqtt.Client
	ftpClient    *ftp.Client
	cameraClient *camera.Client
	logger       *slog.Logger
}

func NewPrinter(config *PrinterConfig) *Printer {
	logger := config.Logger
	if logger == nil {
		logger = slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{
			Level: slog.LevelInfo,
		}))
	}
	logger.Info("Initializing printer", "host", config.Host, "serial", config.SerialNumber)

	return &Printer{
		ipAddr:     config.Host,
		accessCode: config.AccessCode,
		serial:     config.SerialNumber,
		logger:     logger,
		mqttClient: mqtt.NewClient(&mqtt.ClientConfig{
			Host:       config.Host,
			Port:       8883,
			Serial:     config.SerialNumber,
			Username:   "bblp",
			AccessCode: config.AccessCode,
			Timeout:    5 * time.Second,
		}),
		ftpClient: ftp.NewClient(&ftp.ClientConfig{
			Host:       config.Host,
			Port:       990,
			Username:   "bblp",
			AccessCode: config.AccessCode,
		}),
		cameraClient: camera.NewClient(&camera.ClientConfig{
			Hostname:   config.Host,
			AccessCode: config.AccessCode,
			Username:   "bblp",
			Port:       6000,
		}),
	}
}

func (p *Printer) Connect() error {
	p.logger.Info("Connecting to printer", "serial", p.serial)
	if err := p.mqttClient.Connect(); err != nil {
		p.logger.Error("Failed to connect MQTT", "error", err)
		return fmt.Errorf("mqttClient.Connect() error %w", err)
	}
	if err := p.ftpClient.Connect(); err != nil {
		p.logger.Error("Failed to connect FTP", "error", err)
		return fmt.Errorf("ftpClient.Connect() error %w", err)
	}
	p.logger.Info("Printer connected successfully")
	return nil
}

func (p *Printer) Disconnect() error {
	p.logger.Info("Disconnecting from printer", "serial", p.serial)
	p.mqttClient.Disconnect()
	if err := p.ftpClient.Disconnect(); err != nil {
		p.logger.Error("Failed to disconnect FTP", "error", err)
		return fmt.Errorf("ftpClient.Disconnect() error %w", err)
	}
	p.logger.Info("Printer disconnected")
	return nil
}

func (p *Printer) ConnectCamera() error {
	p.logger.Info("Connecting to camera")
	if err := p.cameraClient.Connect(); err != nil {
		p.logger.Error("Failed to connect camera", "error", err)
		return fmt.Errorf("cameraClient.Connect() error %w", err)
	}
	p.logger.Info("Camera connected")
	return nil
}

func (p *Printer) DisconnectCamera() error {
	p.logger.Info("Disconnecting camera")
	if err := p.cameraClient.Disconnect(); err != nil {
		p.logger.Error("Failed to disconnect camera", "error", err)
		return fmt.Errorf("cameraClient.Disconnect() error %w", err)
	}
	p.logger.Info("Camera disconnected")
	return nil
}

func unsafeParseFloat(s string) float64 {
	f, _ := strconv.ParseFloat(s, 64)
	return f
}

func unsafeParseInt(s string) int {
	i, _ := strconv.Atoi(s)
	return i
}

func (p *Printer) Data() (Data, error) {
	p.logger.Info("Fetching printer data")
	data := p.mqttClient.Data()

	final := Data{
		Ams:                     make([]Ams, 0),
		AmsExists:               data.Print.Ams.AmsExistBits == "1",
		BedTargetTemperature:    data.Print.BedTargetTemper,
		BedTemperature:          data.Print.BedTemper,
		AuxiliaryFanSpeed:       unsafeParseInt(data.Print.BigFan1Speed),
		ChamberFanSpeed:         unsafeParseInt(data.Print.BigFan2Speed),
		PartFanSpeed:            unsafeParseInt(data.Print.CoolingFanSpeed),
		HeatbreakFanSpeed:       unsafeParseInt(data.Print.HeatbreakFanSpeed),
		ChamberTemperature:      data.Print.ChamberTemper,
		GcodeFile:               data.Print.GcodeFile,
		GcodeFilePreparePercent: unsafeParseInt(data.Print.GcodeFilePreparePercent),
		GcodeState:              state.GcodeState(data.Print.GcodeState),
		PrintPercentDone:        data.Print.McPercent,
		PrintErrorCode:          data.Print.McPrintErrorCode,
		RemainingPrintTime:      data.Print.McRemainingTime,
		HMS:                     data.Print.HMS,
		NozzleDiameter:          data.Print.NozzleDiameter,
		NozzleTargetTemperature: data.Print.NozzleTargetTemper,
		NozzleTemperature:       data.Print.NozzleTemper,
		Sdcard:                  data.Print.Sdcard,
		WifiSignal:              data.Print.WifiSignal,
	}

	for _, light := range data.Print.LightsReport {
		final.LightReport = append(final.LightReport, LightReport{
			Node: _light.Light(light.Node),
			Mode: _light.Mode(light.Mode),
		})
	}

	colors := make([]color.RGBA, 0)
	for _, col := range data.Print.VtTray.Cols {
		if col == "" {
			colors = append(colors, color.RGBA{})
		} else {
			c, err := parseHexColorFast(col)
			if err != nil {
				p.logger.Error("Failed to parse tray color", "color", col, "error", err)
				return Data{}, fmt.Errorf("parseHexColorFast() error %w", err)
			}
			colors = append(colors, c)
		}
	}

	var trayColor color.RGBA
	if data.Print.VtTray.TrayColor != "" {
		var err error
		trayColor, err = parseHexColorFast(data.Print.VtTray.TrayColor)
		if err != nil {
			p.logger.Error("Failed to parse tray color", "color", data.Print.VtTray.TrayColor, "error", err)
			return Data{}, fmt.Errorf("parseHexColorFast() error %w", err)
		}
	}

	final.VtTray = Tray{
		ID:                unsafeParseInt(data.Print.VtTray.ID),
		BedTemperature:    unsafeParseFloat(data.Print.VtTray.BedTemp),
		Colors:            colors,
		DryingTemperature: unsafeParseFloat(data.Print.VtTray.DryingTemp),
		DryingTime:        unsafeParseInt(data.Print.VtTray.DryingTime),
		NozzleTempMax:     unsafeParseFloat(data.Print.VtTray.NozzleTempMax),
		NozzleTempMin:     unsafeParseFloat(data.Print.VtTray.NozzleTempMin),
		TrayColor:         trayColor,
		TrayDiameter:      unsafeParseFloat(data.Print.VtTray.TrayDiameter),
		TraySubBrands:     data.Print.VtTray.TraySubBrands,
		TrayType:          data.Print.VtTray.TrayType,
		TrayWeight:        unsafeParseInt(data.Print.VtTray.TrayWeight),
	}

	for _, ams := range data.Print.Ams.Ams {
		trays := make([]Tray, 0)
		for _, tray := range ams.Tray {
			trayColors := make([]color.RGBA, 0)
			for _, col := range tray.Cols {
				if col == "" {
					trayColors = append(trayColors, color.RGBA{})
					continue
				}
				c, err := parseHexColorFast(col)
				if err != nil {
					p.logger.Error("Failed to parse AMS tray color", "color", col, "error", err)
					return Data{}, fmt.Errorf("parseHexColorFast() error %w", err)
				}
				trayColors = append(trayColors, c)
			}

			var tColor color.RGBA
			if tray.TrayColor != "" {
				var err error
				tColor, err = parseHexColorFast(tray.TrayColor)
				if err != nil {
					p.logger.Error("Failed to parse AMS tray color", "color", tray.TrayColor, "error", err)
					return Data{}, fmt.Errorf("parseHexColorFast() error %w", err)
				}
			}

			trays = append(trays, Tray{
				ID:                unsafeParseInt(tray.ID),
				BedTemperature:    unsafeParseFloat(tray.BedTemp),
				Colors:            trayColors,
				DryingTemperature: unsafeParseFloat(tray.DryingTemp),
				DryingTime:        unsafeParseInt(tray.DryingTime),
				NozzleTempMax:     unsafeParseFloat(tray.NozzleTempMax),
				NozzleTempMin:     unsafeParseFloat(tray.NozzleTempMin),
				TrayColor:         tColor,
				TrayDiameter:      unsafeParseFloat(tray.TrayDiameter),
				TraySubBrands:     tray.TraySubBrands,
				TrayType:          tray.TrayType,
				TrayWeight:        unsafeParseInt(tray.TrayWeight),
			})
		}
		final.Ams = append(final.Ams, Ams{
			Humidity:    unsafeParseInt(ams.Humidity),
			ID:          unsafeParseInt(ams.ID),
			Temperature: unsafeParseFloat(ams.Temp),
			Trays:       trays,
		})
	}

	p.logger.Info("Printer data fetched successfully")
	return final, nil
}

func (p *Printer) GetSerial() string {
	p.logger.Debug("Retrieving printer serial", "serial", p.serial)
	return p.serial
}

func (p *Printer) GetPrinterState() state.GcodeState {
	s := state.GcodeState(p.mqttClient.Data().Print.GcodeState)
	p.logger.Debug("Retrieved printer state", "state", s.String())
	return s
}

func (p *Printer) GetHMSErrors() []hms.Error {
	errors := p.mqttClient.Data().Print.HMS
	p.logger.Debug("Retrieved HMS errors", "count", len(errors))
	return errors
}

func (p *Printer) setLight(light _light.Light, mode _light.Mode) error {
	p.logger.Info("Setting light", "light", light.String(), "mode", mode.String())
	command := mqtt.NewCommand(mqtt.System).
		AddCommandField("ledctrl").
		AddField("led_node", light).
		AddField("led_mode", mode).
		AddField("led_on_time", 500).
		AddField("led_off_time", 500).
		AddField("loop_times", 1).
		AddField("interval_time", 1000)

	if err := p.mqttClient.Publish(command); err != nil {
		p.logger.Error("Failed to set light", "light", light.String(), "mode", mode.String(), "error", err)
		return fmt.Errorf("error setting light %s: %w", light, err)
	}

	p.logger.Info("Light set successfully", "light", light.String(), "mode", mode.String())
	return nil
}

func (p *Printer) setLightFlashing(light _light.Light, mode _light.Mode, onTime, offTime, loopTimes, intervalTime int) error {
	p.logger.Info("Setting flashing light", "light", light.String(), "mode", mode.String(),
		"onTime", onTime, "offTime", offTime, "loopTimes", loopTimes, "intervalTime", intervalTime)
	command := mqtt.NewCommand(mqtt.System).
		AddCommandField("ledctrl").
		AddField("led_node", light).
		AddField("led_mode", mode).
		AddField("led_on_time", onTime).
		AddField("led_off_time", offTime).
		AddField("loop_times", loopTimes).
		AddField("interval_time", intervalTime)

	if err := p.mqttClient.Publish(command); err != nil {
		p.logger.Error("Failed to set flashing light", "light", light.String(), "error", err)
		return fmt.Errorf("error setting light %s: %w", light, err)
	}

	p.logger.Info("Flashing light set successfully", "light", light.String())
	return nil
}

func (p *Printer) LightOn(light _light.Light) error {
	return p.setLight(light, _light.On)
}

func (p *Printer) LightOff(light _light.Light) error {
	return p.setLight(light, _light.Off)
}

func (p *Printer) LightFlashing(light _light.Light, onTime, offTime, loopTimes, intervalTime int) error {
	return p.setLightFlashing(light, _light.Flashing, onTime, offTime, loopTimes, intervalTime)
}

func (p *Printer) StopPrint() error {
	p.logger.Info("Stopping print")
	s := p.GetPrinterState()
	if s == state.IDLE || s == state.UNKNOWN {
		p.logger.Warn("Cannot stop print: invalid state", "state", s.String())
		return fmt.Errorf("error stopping print: %s", s.String())
	}

	command := mqtt.NewCommand(mqtt.Print).AddCommandField("stop")
	if err := p.mqttClient.Publish(command); err != nil {
		p.logger.Error("Failed to stop print", "error", err)
		return fmt.Errorf("error stopping print: %w", err)
	}

	p.logger.Info("Print stopped successfully")
	return nil
}

func (p *Printer) PausePrint() error {
	p.logger.Info("Pausing print")
	s := p.GetPrinterState()
	if s == state.PAUSE || s == state.UNKNOWN {
		p.logger.Warn("Cannot pause print: invalid state", "state", s.String())
		return fmt.Errorf("error pausing print: %s", s.String())
	}

	command := mqtt.NewCommand(mqtt.Print).AddCommandField("pause")
	if err := p.mqttClient.Publish(command); err != nil {
		p.logger.Error("Failed to pause print", "error", err)
		return fmt.Errorf("error pausing print: %w", err)
	}

	p.logger.Info("Print paused successfully")
	return nil
}

func (p *Printer) ResumePrint() error {
	p.logger.Info("Resuming print")
	s := p.GetPrinterState()
	if s == state.RUNNING || s == state.UNKNOWN {
		p.logger.Warn("Cannot resume print: invalid state", "state", s.String())
		return fmt.Errorf("error resuming print: %s", s.String())
	}

	command := mqtt.NewCommand(mqtt.Print).AddCommandField("resume")
	if err := p.mqttClient.Publish(command); err != nil {
		p.logger.Error("Failed to resume print", "error", err)
		return fmt.Errorf("error resuming print: %w", err)
	}

	p.logger.Info("Print resumed successfully")
	return nil
}

func (p *Printer) SendGcode(gcode []string) error {
	p.logger.Info("Sending gcode commands", "count", len(gcode))
	for _, g := range gcode {
		if !isValidGCode(g) {
			p.logger.Error("Invalid gcode line", "gcode", g)
			return fmt.Errorf("invalid gcode: %s", g)
		}

		command := mqtt.NewCommand(mqtt.Print).AddCommandField("gcode_line").AddParamField(g)
		if err := p.mqttClient.Publish(command); err != nil {
			p.logger.Error("Failed to send gcode", "gcode", g, "error", err)
			return fmt.Errorf("error sending gcode line %s: %w", g, err)
		}
		p.logger.Debug("Gcode line sent", "gcode", g)
	}
	p.logger.Info("All gcode commands sent successfully")
	return nil
}

func (p *Printer) SetBedTemperature(temperature int) error {
	p.logger.Info("Setting bed temperature", "temperature", temperature)
	command := mqtt.NewCommand(mqtt.Print).AddCommandField("gcode_line").AddParamField(fmt.Sprintf("M140 S%d", temperature))
	if err := p.mqttClient.Publish(command); err != nil {
		p.logger.Error("Failed to set bed temperature", "error", err)
		return fmt.Errorf("error setting bed temperature: %w", err)
	}
	p.logger.Info("Bed temperature set successfully", "temperature", temperature)
	return nil
}

func (p *Printer) SetNozzleTemperature(temperature int) error {
	p.logger.Info("Setting nozzle temperature", "temperature", temperature)
	command := mqtt.NewCommand(mqtt.Print).AddCommandField("gcode_line").AddParamField(fmt.Sprintf("M104 S%d", temperature))
	if err := p.mqttClient.Publish(command); err != nil {
		p.logger.Error("Failed to set nozzle temperature", "error", err)
		return fmt.Errorf("error setting nozzle temperature: %w", err)
	}
	p.logger.Info("Nozzle temperature set successfully", "temperature", temperature)
	return nil
}

func (p *Printer) Calibrate(levelBed, vibrationCompensation, motorNoiseCancellation bool) error {
	p.logger.Info("Starting calibration", "levelBed", levelBed, "vibrationCompensation", vibrationCompensation, "motorNoiseCancellation", motorNoiseCancellation)
	bitmask := 0
	if levelBed {
		bitmask |= 1 << 1
	}
	if vibrationCompensation {
		bitmask |= 1 << 2
	}
	if motorNoiseCancellation {
		bitmask |= 1 << 3
	}

	command := mqtt.NewCommand(mqtt.Print).AddCommandField("calibration").AddParamField(strconv.Itoa(bitmask))
	if err := p.mqttClient.Publish(command); err != nil {
		p.logger.Error("Failed to calibrate", "error", err)
		return fmt.Errorf("error calibrating: %w", err)
	}
	p.logger.Info("Calibration started successfully", "bitmask", bitmask)
	return nil
}

func (p *Printer) SetPrintSpeed(speed _printspeed.PrintSpeed) error {
	p.logger.Info("Setting print speed", "speed", speed.String())
	command := mqtt.NewCommand(mqtt.Print).AddCommandField("print_speed").AddParamField(speed)
	if err := p.mqttClient.Publish(command); err != nil {
		p.logger.Error("Failed to set print speed", "error", err)
		return fmt.Errorf("error setting print speed: %w", err)
	}
	p.logger.Info("Print speed set successfully", "speed", speed.String())
	return nil
}

func (p *Printer) SetFanSpeed(fan _fan.Fan, speed int) error {
	p.logger.Info("Setting fan speed", "fan", fan.String(), "speed", speed)
	if speed < 0 || speed > 255 {
		p.logger.Error("Invalid fan speed", "speed", speed)
		return fmt.Errorf("invalid speed: %d; must be between 0 and 255", speed)
	}

	command := mqtt.NewCommand(mqtt.Print).AddCommandField("gcode_line").AddParamField(fmt.Sprintf("M106 P%d S%d", fan, speed))
	if err := p.mqttClient.Publish(command); err != nil {
		p.logger.Error("Failed to set fan speed", "fan", fan.String(), "speed", speed, "error", err)
		return fmt.Errorf("error setting fan speed: %w", err)
	}
	p.logger.Info("Fan speed set successfully", "fan", fan.String(), "speed", speed)
	return nil
}

func (p *Printer) StoreFile(path string, file os.File) error {
	p.logger.Info("Storing file", "path", path)
	if err := p.ftpClient.StoreFile(path, file); err != nil {
		p.logger.Error("Failed to store file", "path", path, "error", err)
		return err
	}
	p.logger.Info("File stored successfully", "path", path)
	return nil
}

func (p *Printer) ListDirectory(path string) ([]os.FileInfo, error) {
	p.logger.Info("Listing directory", "path", path)
	infos, err := p.ftpClient.ListDir(path)
	if err != nil {
		p.logger.Error("Failed to list directory", "path", path, "error", err)
		return nil, err
	}
	p.logger.Info("Directory listed successfully", "path", path, "count", len(infos))
	return infos, nil
}

func (p *Printer) RetrieveFile(path string) (os.File, error) {
	p.logger.Info("Retrieving file", "path", path)
	file, err := p.ftpClient.RetrieveFile(path)
	if err != nil {
		p.logger.Error("Failed to retrieve file", "path", path, "error", err)
		return os.File{}, err
	}
	p.logger.Info("File retrieved successfully", "path", path)
	return file, nil
}

func (p *Printer) DeleteFile(path string) error {
	p.logger.Info("Deleting file", "path", path)
	if err := p.ftpClient.DeleteFile(path); err != nil {
		p.logger.Error("Failed to delete file", "path", path, "error", err)
		return err
	}
	p.logger.Info("File deleted successfully", "path", path)
	return nil
}

func (p *Printer) CaptureCameraFrame() ([]byte, error) {
	p.logger.Info("Capturing camera frame", "printer", p.serial)
	frame, err := p.cameraClient.CaptureFrame()
	if err != nil {
		p.logger.Error("Failed to capture camera frame", "printer", p.serial, "error", err)
		return nil, err
	}
	p.logger.Info("Camera frame captured successfully", "printer", p.serial, "frameSize", len(frame))
	return frame, nil
}

func (p *Printer) StartCameraStream() (<-chan []byte, error) {
	p.logger.Info("Starting camera stream", "printer", p.serial)
	stream, err := p.cameraClient.StartStream()
	if err != nil {
		p.logger.Error("Failed to start camera stream", "printer", p.serial, "error", err)
		return nil, err
	}
	p.logger.Info("Camera stream started successfully", "printer", p.serial)
	return stream, nil
}

func (p *Printer) StopCameraStream() error {
	p.logger.Info("Stopping camera stream", "printer", p.serial)
	err := p.cameraClient.StopStream()
	if err != nil {
		p.logger.Error("Failed to stop camera stream", "printer", p.serial, "error", err)
		return err
	}
	p.logger.Info("Camera stream stopped successfully", "printer", p.serial)
	return nil
}
