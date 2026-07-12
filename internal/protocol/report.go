package protocol

type Report struct {
	Print *PrintReport `json:"print,omitempty"`
}

type PrintReport struct {
	HMSErrors  []HMSErrorReport `json:"hms,omitempty"`
	Command    string           `json:"command,omitempty"`
	GcodeState string           `json:"gcode_state,omitempty"`

	// Virtual Tray (external spool)
	VtTray *TrayReport `json:"vt_tray,omitempty"`

	AMS    *AMSReport    `json:"ams,omitempty"`
	Device *DeviceReport `json:"device,omitempty"`
	IPCam  *IPCamReport  `json:"ipcam,omitempty"`
	XCam   *XCamReport   `json:"xcam,omitempty"`
	Online *OnlineReport `json:"online,omitempty"`

	BedTargetTemper float64 `json:"bed_target_temper,omitempty"`
	BedTemper       float64 `json:"bed_temper,omitempty"`

	NozzleTargetTemper float64 `json:"nozzle_target_temper,omitempty"`
	NozzleTemper       float64 `json:"nozzle_temper,omitempty"`
	NozzleDiameter     string  `json:"nozzle_diameter,omitempty"`
	NozzleType         string  `json:"nozzle_type,omitempty"`

	LayerNum        int `json:"layer_num,omitempty"`
	TotalLayerNum   int `json:"total_layer_num,omitempty"`
	MCPercent       int `json:"mc_percent,omitempty"`
	MCRemainingTime int `json:"mc_remaining_time,omitempty"`

	Percent    int `json:"percent,omitempty"`
	RemainTime int `json:"remain_time,omitempty"`

	GcodeFile string `json:"gcode_file,omitempty"`
	File      string `json:"file,omitempty"`

	FailReason string `json:"fail_reason,omitempty"`

	WifiSignal string `json:"wifi_signal,omitempty"`

	// Raw bitfields
	Fun  string `json:"fun,omitempty"`
	Fun2 string `json:"fun2,omitempty"`
	Stat string `json:"stat,omitempty"`
	Aux  string `json:"aux,omitempty"`

	// Misc protocol data
	Stg []int `json:"stg,omitempty"`

	LightsReport []LightsReport `json:"lights_report,omitempty"`

	UpgradeState *UpgradeStateReport `json:"upgrade_state,omitempty"`
	Upload       *UploadReport       `json:"upload,omitempty"`
}

type HMSErrorReport struct {
	Attr   int    `json:"attr,omitempty"`
	Code   int    `json:"code,omitempty"`
	TSBoot int    `json:"ts_boot,omitempty"`
	TSUnix string `json:"ts_unix,omitempty"`
}

type DeviceReport struct {
	Airduct  *AirductReport  `json:"airduct,omitempty"`
	Bed      *TempNode       `json:"bed,omitempty"`
	Cam      *CameraReport   `json:"cam,omitempty"`
	CTC      *TempNode       `json:"ctc,omitempty"`
	Extruder *ExtruderReport `json:"extruder,omitempty"`
	Nozzle   *NozzleReport   `json:"nozzle,omitempty"`
	Plate    *PlateReport    `json:"plate,omitempty"`

	ExtTool *ExtToolReport `json:"ext_tool,omitempty"`

	Fan int `json:"fan,omitempty"`

	Laser *LaserReport `json:"laser,omitempty"`

	Screen *ScreenReport `json:"screen,omitempty"`

	Type int `json:"type,omitempty"`
}

type ExtruderReport struct {
	Info  []ExtruderInfo `json:"info,omitempty"`
	State int            `json:"state,omitempty"`
}

type ExtruderInfo struct {
	ID int `json:"id"`

	Info int `json:"info"`
	Stat int `json:"stat"`

	Temp int `json:"temp"`

	Snow int `json:"snow"`
	Hnow int `json:"hnow"`

	Star int `json:"star"`
	Htar int `json:"htar"`

	Spre int `json:"spre"`
	Hpre int `json:"hpre"`

	FilamBak []any `json:"filam_bak"`
}

type TempNode struct {
	Info  TempInfo `json:"info,omitempty"`
	State int      `json:"state,omitempty"`
}

type TempInfo struct {
	Temp int `json:"temp,omitempty"`
}

type AirductReport struct {
	ModeCur     int           `json:"modeCur,omitempty"`
	ModeFunc    int           `json:"modeFunc,omitempty"`
	ModeList    []AirductMode `json:"modeList,omitempty"`
	ModeVisable int           `json:"modeVisable,omitempty"`

	Parts []AirductPart `json:"parts,omitempty"`

	SubFunc    int `json:"subFunc,omitempty"`
	SubMode    int `json:"subMode,omitempty"`
	SubVisable int `json:"subVisable,omitempty"`

	Version int `json:"version,omitempty"`
}

type AirductMode struct {
	Ctrl   []int `json:"ctrl,omitempty"`
	ModeID int   `json:"modeId,omitempty"`
	Off    []int `json:"off,omitempty"`
}

type AirductPart struct {
	Func     int `json:"func,omitempty"`
	ID       int `json:"id,omitempty"`
	Range    int `json:"range,omitempty"`
	State    int `json:"state,omitempty"`
	TarState int `json:"tar_state,omitempty"`
}

type CameraReport struct {
	Laser LaserCameraReport `json:"laser,omitempty"`

	TimelapsePath string `json:"timelapse_path,omitempty"`

	TLExternalFreeKB  int `json:"tl_external_free_kb,omitempty"`
	TLExternalTotalKB int `json:"tl_external_total_kb,omitempty"`

	TLInternalFreeKB  int `json:"tl_internal_free_kb,omitempty"`
	TLInternalTotalKB int `json:"tl_internal_total_kb,omitempty"`
}

type LaserCameraReport struct {
	Cond  int `json:"cond,omitempty"`
	State int `json:"state,omitempty"`
}

type ExtToolReport struct {
	Calib   int    `json:"calib,omitempty"`
	LowPrec bool   `json:"low_prec,omitempty"`
	Mount   int    `json:"mount,omitempty"`
	Mount3D int    `json:"mount_3d,omitempty"`
	ThTemp  int    `json:"th_temp,omitempty"`
	Type    string `json:"type,omitempty"`
}

type LaserReport struct {
	Power int `json:"power,omitempty"`
}

type ScreenReport struct {
	Backlight int `json:"backlight,omitempty"`
}

type PlateReport struct {
	Base     int    `json:"base,omitempty"`
	Cali2DID string `json:"cali2d_id,omitempty"`
	CurID    string `json:"cur_id,omitempty"`
	Mat      int    `json:"mat,omitempty"`
	TarID    string `json:"tar_id,omitempty"`
}

type NozzleReport struct {
	Exist int `json:"exist,omitempty"`
	HC    int `json:"hc,omitempty"`

	Info []NozzleInfo `json:"info,omitempty"`

	SrcID int `json:"src_id,omitempty"`
	State int `json:"state,omitempty"`
	TarID int `json:"tar_id,omitempty"`
}

type NozzleInfo struct {
	ColorM   string  `json:"color_m,omitempty"`
	Diameter float64 `json:"diameter,omitempty"`

	FilaID string `json:"fila_id,omitempty"`

	ID int `json:"id,omitempty"`

	PT int `json:"p_t,omitempty"`

	SN string `json:"sn,omitempty"`

	Stat int `json:"stat,omitempty"`

	TM int `json:"tm,omitempty"`

	Type string `json:"type,omitempty"`

	Wear float64 `json:"wear,omitempty"`
}

type AMSReport struct {
	AMS []AMSUnitReport `json:"ams,omitempty"`

	AMSExistBits    string `json:"ams_exist_bits,omitempty"`
	AMSExistBitsRaw string `json:"ams_exist_bits_raw,omitempty"`

	CaliID   int `json:"cali_id,omitempty"`
	CaliStat int `json:"cali_stat,omitempty"`

	CFS []int `json:"cfs,omitempty"`

	InsertFlag  bool `json:"insert_flag,omitempty"`
	PowerOnFlag bool `json:"power_on_flag,omitempty"`

	TrayExistBits    string `json:"tray_exist_bits,omitempty"`
	TrayHallOutBits  string `json:"tray_hall_out_bits,omitempty"`
	TrayIsBblBits    string `json:"tray_is_bbl_bits,omitempty"`
	TrayNow          string `json:"tray_now,omitempty"`
	TrayPre          string `json:"tray_pre,omitempty"`
	TrayReadDoneBits string `json:"tray_read_done_bits,omitempty"`
	TrayReadingBits  string `json:"tray_reading_bits,omitempty"`
	TrayTar          string `json:"tray_tar,omitempty"`

	UnbindAMSStat int `json:"unbind_ams_stat,omitempty"`

	Version int `json:"version,omitempty"`
}

type AMSUnitReport struct {
	DrySetting *DrySetting `json:"dry_setting,omitempty"`

	DrySFReason []any `json:"dry_sf_reason,omitempty"`

	DryTime int `json:"dry_time,omitempty"`

	Humidity    string `json:"humidity,omitempty"`
	HumidityRaw string `json:"humidity_raw,omitempty"`

	ID string `json:"id,omitempty"`

	Info string `json:"info,omitempty"`

	Temp string `json:"temp,omitempty"`

	Tray []TrayReport `json:"tray,omitempty"`
}

type DrySetting struct {
	DryDuration    int    `json:"dry_duration,omitempty"`
	DryFilament    string `json:"dry_filament,omitempty"`
	DryTemperature int    `json:"dry_temperature,omitempty"`
}

type TrayReport struct {
	ID string `json:"id,omitempty"`

	BedTemp     string `json:"bed_temp,omitempty"`
	BedTempType string `json:"bed_temp_type,omitempty"`

	CaliIdx int `json:"cali_idx,omitempty"`

	Cols []string `json:"cols,omitempty"`

	CType int `json:"ctype,omitempty"`

	DryingTemp string `json:"drying_temp,omitempty"`
	DryingTime string `json:"drying_time,omitempty"`

	NozzleTempMax string `json:"nozzle_temp_max,omitempty"`
	NozzleTempMin string `json:"nozzle_temp_min,omitempty"`

	Remaining      int  `json:"remain,omitempty"`
	RemainingGrams *int `json:"remain_g,omitempty"`

	State int `json:"state,omitempty"`

	TagUID string `json:"tag_uid,omitempty"`

	TotalLen int `json:"total_len,omitempty"`

	TrayColor string `json:"tray_color,omitempty"`

	TrayDiameter string `json:"tray_diameter,omitempty"`

	TrayIDName string `json:"tray_id_name,omitempty"`

	TrayInfoIdx string `json:"tray_info_idx,omitempty"`

	TraySubBrands string `json:"tray_sub_brands,omitempty"`

	TrayType string `json:"tray_type,omitempty"`

	TrayUUID string `json:"tray_uuid,omitempty"`

	TrayWeight string `json:"tray_weight,omitempty"`

	XCamInfo string `json:"xcam_info,omitempty"`
}

type JobReport struct {
	CurStage *JobStageState `json:"cur_stage,omitempty"`

	JobState int `json:"job_state,omitempty"`

	Stage []JobStage `json:"stage,omitempty"`
}

type JobStageState struct {
	Idx   int `json:"idx,omitempty"`
	State int `json:"state,omitempty"`
}

type JobStage struct {
	ClockIn bool `json:"clock_in,omitempty"`

	Color []string `json:"color,omitempty"`

	Diameter []float64 `json:"diameter,omitempty"`

	EstTime int `json:"est_time,omitempty"`

	Heigh float64 `json:"heigh,omitempty"`

	IDX int `json:"idx,omitempty"`

	Platform string `json:"platform,omitempty"`

	PrintThen bool `json:"print_then,omitempty"`

	ProcList []any `json:"proc_list,omitempty"`

	Tool []string `json:"tool,omitempty"`

	Type int `json:"type,omitempty"`
}

type LightsReport struct {
	Mode string `json:"mode,omitempty"`
	Node string `json:"node,omitempty"`
}

type UpgradeStateReport struct {
	AHBNewVersionNumber string `json:"ahb_new_version_number,omitempty"`
	AMSNewVersionNumber string `json:"ams_new_version_number,omitempty"`
	ExtNewVersionNumber string `json:"ext_new_version_number,omitempty"`

	ConsistencyRequest bool `json:"consistency_request,omitempty"`

	DisState int `json:"dis_state,omitempty"`
	ErrCode  int `json:"err_code,omitempty"`

	ForceUpgrade bool `json:"force_upgrade,omitempty"`

	Idx  int `json:"idx,omitempty"`
	Idx2 int `json:"idx2,omitempty"`

	LowerLimit string `json:"lower_limit,omitempty"`

	Message string `json:"message,omitempty"`

	Module string `json:"module,omitempty"`

	NewVersionState int `json:"new_version_state,omitempty"`

	OTANewVersionNumber string `json:"ota_new_version_number,omitempty"`

	Progress string `json:"progress,omitempty"`

	Rate string `json:"rate,omitempty"`

	SequenceID int `json:"sequence_id,omitempty"`

	SN string `json:"sn,omitempty"`

	Status string `json:"status,omitempty"`
}

type UploadReport struct {
	FileSize int `json:"file_size,omitempty"`

	FinishSize int `json:"finish_size,omitempty"`

	Message string `json:"message,omitempty"`

	OssURL string `json:"oss_url,omitempty"`

	Progress int `json:"progress,omitempty"`

	SequenceID string `json:"sequence_id,omitempty"`

	Speed int `json:"speed,omitempty"`

	Status string `json:"status,omitempty"`

	TaskID string `json:"task_id,omitempty"`

	TimeRemaining int `json:"time_remaining,omitempty"`

	TroubleID string `json:"trouble_id,omitempty"`
}

type IPCamReport struct {
	AgoraService string `json:"agora_service,omitempty"`

	BRTCService string `json:"brtc_service,omitempty"`

	BSState int `json:"bs_state,omitempty"`

	CapPicEnable string `json:"cap_pic_enable,omitempty"`

	IPCamDev string `json:"ipcam_dev,omitempty"`

	IPCamRecord string `json:"ipcam_record,omitempty"`

	LaserPreviewRes int `json:"laser_preview_res,omitempty"`

	LiveviewPreview bool `json:"liveview_preview,omitempty"`

	ModeBits int `json:"mode_bits,omitempty"`

	Resolution string `json:"resolution,omitempty"`

	RTSPURL string `json:"rtsp_url,omitempty"`

	Timelapse string `json:"timelapse,omitempty"`

	TLExternalFreeKB int `json:"tl_external_free_kb,omitempty"`

	TLExternalTotalKB int `json:"tl_external_total_kb,omitempty"`

	TLInternalFreeKB int `json:"tl_internal_free_kb,omitempty"`

	TLInternalTotalKB int `json:"tl_internal_total_kb,omitempty"`

	TLStoreHPDType int `json:"tl_store_hpd_type,omitempty"`

	TLStorePathType int `json:"tl_store_path_type,omitempty"`

	TUTKServer string `json:"tutk_server,omitempty"`
}

type XCamReport struct {
	AllowSkipParts bool `json:"allow_skip_parts,omitempty"`

	BuildplateMarkerDetector bool `json:"buildplate_marker_detector,omitempty"`

	Cfg int `json:"cfg,omitempty"`

	FirstLayerInspector bool `json:"first_layer_inspector,omitempty"`

	HaltPrintSensitivity string `json:"halt_print_sensitivity,omitempty"`

	PrintHalt bool `json:"print_halt,omitempty"`

	PrintingMonitor bool `json:"printing_monitor,omitempty"`

	SpaghettiDetector bool `json:"spaghetti_detector,omitempty"`
}

type OnlineReport struct {
	AHB bool `json:"ahb,omitempty"`

	RFID bool `json:"rfid,omitempty"`

	Version int `json:"version,omitempty"`
}
