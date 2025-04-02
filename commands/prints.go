package commands

import (
	"fmt"

	"github.com/torbenconto/bambulabs_api/internal/mqtt"
	"github.com/torbenconto/bambulabs_api/printspeed"
	"github.com/torbenconto/bambulabs_api/state"
)

type PrintsUpdateBlock struct {
	GcodeFile               string `json:"gcode_file"`
	GcodeFilePreparePercent string `json:"gcode_file_prepare_percent"`
	GcodeStartTime          string `json:"gcode_start_time"`
	GcodeState              string `json:"gcode_state"`
	McPercent               int    `json:"mc_percent"`
	McPrintErrorCode        string `json:"mc_print_error_code"`
	McPrintStage            string `json:"mc_print_stage"`
	McPrintSubStage         int    `json:"mc_print_sub_stage"`
	McRemainingTime         int    `json:"mc_remaining_time"`
	PrintError              int    `json:"print_error"`
	PrintGcodeAction        int    `json:"print_gcode_action"`
	PrintRealAction         int    `json:"print_real_action"`
	PrintType               string `json:"print_type"`
}

type Prints struct {
	mqttClient *mqtt.Client
	PrintJob   *PrintJob
}

type PrintJob struct {
	mqttClient     *mqtt.Client
	printsBlock    *PrintsUpdateBlock
	oldPrintsBlock *PrintsUpdateBlock
}

func CreatePrintsInstance(mqttClient *mqtt.Client) *Prints {
	printJob := &PrintJob{
		mqttClient:  mqttClient,
		printsBlock: &PrintsUpdateBlock{},
	}
	prints := &Prints{
		mqttClient: mqttClient,
		PrintJob:   printJob,
	}
	mqttClient.OnUpdate(prints.handleUpdate)
	return prints
}

func (p *Prints) Stop() error {
	command := mqtt.NewCommand(mqtt.Print).AddCommandField("stop")

	if err := p.mqttClient.Publish(command); err != nil {
		return fmt.Errorf("error stopping print: %w", err)
	}

	return nil
}

// StopPrint fully stops the current print job.
// Function works independently but problems exist with the underlying.
func (p *Prints) Pause() error {
	command := mqtt.NewCommand(mqtt.Print).AddCommandField("pause")

	if err := p.mqttClient.Publish(command); err != nil {
		return fmt.Errorf("error pausing print: %w", err)
	}

	return nil
}

// PausePrint pauses the current print job.
// Function works independently but problems exist with the underlying.
func (p *Prints) PausePrint() error {
	command := mqtt.NewCommand(mqtt.Print).AddCommandField("pause")

	if err := p.mqttClient.Publish(command); err != nil {
		return fmt.Errorf("error pausing print: %w", err)
	}

	return nil
}

// ResumePrint resumes a paused print job.
// Function works independently but problems exist with the underlying.
func (p *Prints) ResumePrint() error {
	command := mqtt.NewCommand(mqtt.Print).AddCommandField("resume")

	if err := p.mqttClient.Publish(command); err != nil {
		return fmt.Errorf("error resuming print: %w", err)
	}

	return nil
}

// Print3mfFile prints a ".gcode.3mf" file which resides on the printer. A file url (beginning with ftp:/// or file:///) should be passed in.
// You can upload a file through the ftp store function, then print it with this function using the url ftp:///[filename]. Make sure that it ends in .gcode or .gcode.3mf.
// The plate number should almost always be 1.
// This function is working and has been tested on:
// - [x] X1 Carbon
// - [ ] P1S (not tested)
func (p *Prints) Start3MFPrint(fileUrl string, plate int, useAms bool, timelapse bool, calibrate bool, inspectLayers bool) error {
	command := mqtt.NewCommand(mqtt.Print).AddCommandField("project_file").AddParamField(fmt.Sprintf("Metadata/plate_%d.gcode", plate))

	command.AddField("project_id", "0")
	command.AddField("profile_id", "0")
	command.AddField("task_id", "0")
	command.AddField("subtask_id", "0")
	command.AddField("subtask_name", "")
	command.AddField("file", "")
	command.AddField("url", fileUrl)
	command.AddField("md5", "")
	command.AddField("timelapse", timelapse)
	command.AddField("bed_type", "auto")
	command.AddField("bed_levelling", calibrate)
	command.AddField("flow_cali", calibrate)
	command.AddField("vibration_cali", calibrate)
	command.AddField("layer_inspect", inspectLayers)
	command.AddField("ams_mapping", "")
	command.AddField("use_ams", useAms)

	if err := p.mqttClient.Publish(command); err != nil {
		return fmt.Errorf("error printing %s: %w", fileUrl, err)
	}

	return nil
}

// SetPrintSpeed sets the print speed of the printer.
func (p *Prints) SetPrintSpeed(speed printspeed.PrintSpeed) error {
	command := mqtt.NewCommand(mqtt.Print).AddCommandField("print_speed").AddParamField(speed)

	if err := p.mqttClient.Publish(command); err != nil {
		return fmt.Errorf("error setting print speed: %w", err)
	}

	return nil
}

func (pj *PrintJob) GetGCodeFile() string {
	return pj.printsBlock.GcodeFile
}

func (pj *PrintJob) GetGCodeFilePreparePercent() string {
	return pj.printsBlock.GcodeFilePreparePercent
}

func (pj *PrintJob) GetGCodeStartTime() string {
	return pj.printsBlock.GcodeStartTime
}

func (pj *PrintJob) GetGCodeState() string {
	return pj.printsBlock.GcodeState
}

func (pj *PrintJob) GetMcPercent() int {
	return pj.printsBlock.McPercent
}

func (pj *PrintJob) GetMcPrintErrorCode() string {
	return pj.printsBlock.McPrintErrorCode
}

func (pj *PrintJob) GetMcPrintStage() string {
	return pj.printsBlock.McPrintStage
}

func (pj *PrintJob) GetMcPrintSubStage() int {
	return pj.printsBlock.McPrintSubStage
}

func (pj *PrintJob) GetMcRemainingTime() int {
	return pj.printsBlock.McRemainingTime
}

func (pj *PrintJob) GetPrintError() int {
	return pj.printsBlock.PrintError
}

func (pj *PrintJob) GetPrintGcodeAction() int {
	return pj.printsBlock.PrintGcodeAction
}

func (pj *PrintJob) GetPrintRealAction() int {
	return pj.printsBlock.PrintRealAction
}

func (pj *PrintJob) GetPrintType() string {
	return pj.printsBlock.PrintType
}

// handleUpdate processes updates from the MQTT client and sends them to the update channel.
func (p *Prints) handleUpdate(Data mqtt.Message) {
	p.PrintJob.oldPrintsBlock = p.PrintJob.printsBlock

	// Assuming the message contains the necessary fields to populate PrintsUpdateBlock
	p.PrintJob.printsBlock.GcodeFile = Data.Print.GcodeFile
	if p.PrintJob.printsBlock.GcodeFile != p.PrintJob.oldPrintsBlock.GcodeFile {
		Emit(PrintStarted)
	}

	p.PrintJob.printsBlock.GcodeFilePreparePercent = Data.Print.GcodeFilePreparePercent
	p.PrintJob.printsBlock.GcodeStartTime = Data.Print.GcodeStartTime
	p.PrintJob.printsBlock.GcodeState = Data.Print.GcodeState
	if p.PrintJob.printsBlock.GcodeState != p.PrintJob.oldPrintsBlock.GcodeState {
		if state.GcodeState(p.PrintJob.printsBlock.GcodeState) == state.PAUSE {
			Emit(PrintPaused)
		}
		if state.GcodeState(p.PrintJob.printsBlock.GcodeState) == state.FINISH {
			Emit(PrintFinished)
		}
		if state.GcodeState(p.PrintJob.printsBlock.GcodeState) == state.FAILED {
			Emit(PrintFailed)
		}
	}

	p.PrintJob.printsBlock.McPercent = Data.Print.McPercent
	if p.PrintJob.printsBlock.McPercent != p.PrintJob.oldPrintsBlock.McPercent {
		Emit(PrintPercent)
	}

	p.PrintJob.printsBlock.McPrintErrorCode = Data.Print.McPrintErrorCode
	if p.PrintJob.printsBlock.McPrintErrorCode != p.PrintJob.oldPrintsBlock.McPrintErrorCode && p.PrintJob.printsBlock.McPrintErrorCode != "" {
		Emit(PrintError)
	}

	p.PrintJob.printsBlock.McPrintStage = Data.Print.McPrintStage
	p.PrintJob.printsBlock.McPrintSubStage = Data.Print.McPrintSubStage
	p.PrintJob.printsBlock.McRemainingTime = Data.Print.McRemainingTime
	if p.PrintJob.printsBlock.McRemainingTime != p.PrintJob.oldPrintsBlock.McRemainingTime {
		Emit(PrintTime)
	}

	p.PrintJob.printsBlock.PrintError = Data.Print.PrintError
	if p.PrintJob.printsBlock.PrintError != p.PrintJob.oldPrintsBlock.PrintError && p.PrintJob.printsBlock.PrintError != 0 {
		Emit(PrintError)
	}

	p.PrintJob.printsBlock.PrintGcodeAction = Data.Print.PrintGcodeAction
	p.PrintJob.printsBlock.PrintRealAction = Data.Print.PrintRealAction
	p.PrintJob.printsBlock.PrintType = Data.Print.PrintType
}
