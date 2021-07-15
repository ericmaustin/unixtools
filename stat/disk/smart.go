package disk

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"gopkg.in/yaml.v2"
	"os/exec"
)

var ErrSmartctlNotInstalled = errors.New("smartctl is not installed on this system")

type SMARTExitCode int

const (
	SmartOK                SMARTExitCode = 0
	SmartCmdFailed                       = 1
	DeviceOpenFailed                     = 2
	SmartResponseError                   = 4
	SmartDiskFailing                     = 8
	SmartPrefail                         = 16
	SmartPreviousPrefail                 = 32
	SmartErrorLogHasErrors               = 64
	SmartSelfTestErrors                  = 128
)

var (
	smartExitCodeString = map[SMARTExitCode]string{
		SmartOK:                "OK",
		SmartCmdFailed:         "Command failed",
		DeviceOpenFailed:       "Device open failed, device did not return an IDENTIFY DEVICE structure, or device is in a low-power mode",
		SmartResponseError:     "Some SMART or other ATA command to the disk failed, or there was a checksum error in a SMART data structure",
		SmartDiskFailing:       "DISK FAILING",
		SmartPrefail:           "prefail Attributes <= threshold",
		SmartPreviousPrefail:   "SMART status check returned \"DISK OK\" but we found that some (usage or prefail) Attributes have been <= threshold at some time in the past",
		SmartErrorLogHasErrors: "The device error log contains records of errors",
		SmartSelfTestErrors:    "The device self-test log contains records of errors",
	}
)

func (ec SMARTExitCode) String() string {
	return smartExitCodeString[ec]
}

func (ec SMARTExitCode) MarshalJSON() ([]byte, error) {
	return []byte(ec.String()), nil
}

type SMARTInfo struct {
	ExitCode SMARTExitCode `yaml:"exit_code" json:"exit_code"`
	Smartctl struct {
		Version      []int    `yaml:"version" json:"version"`
		SvnRevision  string   `yaml:"svn_revision" json:"svn_revision"`
		PlatformInfo string   `yaml:"platform_info" json:"platform_info"`
		BuildInfo    string   `yaml:"build_info" json:"build_info"`
		Argv         []string `yaml:"argv" json:"argv"`
		ExitStatus   int      `yaml:"exit_status" json:"exit_status"`
	} `yaml:"smartctl" json:"smartctl"`
	Device struct {
		Name     string `yaml:"name" json:"name"`
		InfoName string `yaml:"info_name" json:"info_name"`
		Type     string `yaml:"type" json:"type"`
		Protocol string `yaml:"protocol" json:"protocol"`
	} `yaml:"device" json:"device"`
	ModelFamily  string `yaml:"model_family" json:"model_family"`
	ModelName    string `yaml:"model_name" json:"model_name"`
	SerialNumber string `yaml:"serial_number" json:"serial_number"`
	Wwn          struct {
		Naa int   `yaml:"naa" json:"naa"`
		Oui int   `yaml:"oui" json:"oui"`
		ID  int64 `yaml:"id" json:"id"`
	} `yaml:"wwn" json:"wwn"`
	FirmwareVersion string `yaml:"firmware_version" json:"firmware_version"`
	UserCapacity    struct {
		Blocks int64 `yaml:"blocks" json:"blocks"`
		Bytes  int64 `yaml:"bytes" json:"bytes"`
	} `yaml:"user_capacity" json:"user_capacity"`
	LogicalBlockSize  int `yaml:"logical_block_size" json:"logical_block_size"`
	PhysicalBlockSize int `yaml:"physical_block_size" json:"physical_block_size"`
	RotationRate      int `yaml:"rotation_rate" json:"rotation_rate"`
	FormFactor        struct {
		AtaValue int    `yaml:"ata_value" json:"ata_value"`
		Name     string `yaml:"name" json:"name"`
	} `yaml:"form_factor" json:"form_factor"`
	InSmartctlDatabase bool `yaml:"in_smartctl_database" json:"in_smartctl_database"`
	AtaVersion         struct {
		String     string `yaml:"string" json:"string"`
		MajorValue int    `yaml:"major_value" json:"major_value"`
		MinorValue int    `yaml:"minor_value" json:"minor_value"`
	} `yaml:"ata_version" json:"ata_version"`
	SataVersion struct {
		String string `yaml:"string" json:"string"`
		Value  int    `yaml:"value" json:"value"`
	} `yaml:"sata_version" json:"sata_version"`
	InterfaceSpeed struct {
		Max struct {
			SataValue      int    `yaml:"sata_value" json:"sata_value"`
			String         string `yaml:"string" json:"string"`
			UnitsPerSecond int    `yaml:"units_per_second" json:"units_per_second"`
			BitsPerUnit    int    `yaml:"bits_per_unit" json:"bits_per_unit"`
		} `yaml:"max" json:"max"`
		Current struct {
			SataValue      int    `yaml:"sata_value" json:"sata_value"`
			String         string `yaml:"string" json:"string"`
			UnitsPerSecond int    `yaml:"units_per_second" json:"units_per_second"`
			BitsPerUnit    int    `yaml:"bits_per_unit" json:"bits_per_unit"`
		} `yaml:"current" json:"current"`
	} `yaml:"interface_speed" json:"interface_speed"`
	LocalTime struct {
		TimeT   int    `yaml:"time_t" json:"time_t"`
		Asctime string `yaml:"asctime" json:"asctime"`
	} `yaml:"local_time" json:"local_time"`
	SmartStatus struct {
		Passed bool `yaml:"passed" json:"passed"`
	} `yaml:"smart_status" json:"smart_status"`
	AtaSmartData struct {
		OfflineDataCollection struct {
			Status struct {
				Value  int    `yaml:"value" json:"value"`
				String string `yaml:"string" json:"string"`
				Passed bool   `yaml:"passed" json:"passed"`
			} `yaml:"status" json:"status"`
			CompletionSeconds int `yaml:"completion_seconds" json:"completion_seconds"`
		} `yaml:"offline_data_collection" json:"offline_data_collection"`
		SelfTest struct {
			Status struct {
				Value  int    `yaml:"value" json:"value"`
				String string `yaml:"string" json:"string"`
				Passed bool   `yaml:"passed" json:"passed"`
			} `yaml:"status" json:"status"`
			PollingMinutes struct {
				Short      int `yaml:"short" json:"short"`
				Extended   int `yaml:"extended" json:"extended"`
				Conveyance int `yaml:"conveyance" json:"conveyance"`
			} `yaml:"polling_minutes" json:"polling_minutes"`
		} `yaml:"self_test" json:"self_test"`
		Capabilities struct {
			Values                        []int `yaml:"values" json:"values"`
			ExecOfflineImmediateSupported bool  `yaml:"exec_offline_immediate_supported" json:"exec_offline_immediate_supported"`
			OfflineIsAbortedUponNewCmd    bool  `yaml:"offline_is_aborted_upon_new_cmd" json:"offline_is_aborted_upon_new_cmd"`
			OfflineSurfaceScanSupported   bool  `yaml:"offline_surface_scan_supported" json:"offline_surface_scan_supported"`
			SelfTestsSupported            bool  `yaml:"self_tests_supported" json:"self_tests_supported"`
			ConveyanceSelfTestSupported   bool  `yaml:"conveyance_self_test_supported" json:"conveyance_self_test_supported"`
			SelectiveSelfTestSupported    bool  `yaml:"selective_self_test_supported" json:"selective_self_test_supported"`
			AttributeAutosaveEnabled      bool  `yaml:"attribute_autosave_enabled" json:"attribute_autosave_enabled"`
			ErrorLoggingSupported         bool  `yaml:"error_logging_supported" json:"error_logging_supported"`
			GpLoggingSupported            bool  `yaml:"gp_logging_supported" json:"gp_logging_supported"`
		} `yaml:"capabilities" json:"capabilities"`
	} `yaml:"ata_smart_data" json:"ata_smart_data"`
	AtaSctCapabilities struct {
		Value                         int  `yaml:"value" json:"value"`
		ErrorRecoveryControlSupported bool `yaml:"error_recovery_control_supported" json:"error_recovery_control_supported"`
		FeatureControlSupported       bool `yaml:"feature_control_supported" json:"feature_control_supported"`
		DataTableSupported            bool `yaml:"data_table_supported" json:"data_table_supported"`
	} `yaml:"ata_sct_capabilities" json:"ata_sct_capabilities"`
	AtaSmartAttributes struct {
		Revision int `yaml:"revision" json:"revision"`
		Table    []struct {
			ID         int    `yaml:"id" json:"id"`
			Name       string `yaml:"name" json:"name"`
			Value      int    `yaml:"value" json:"value"`
			Worst      int    `yaml:"worst" json:"worst"`
			Thresh     int    `yaml:"thresh" json:"thresh"`
			WhenFailed string `yaml:"when_failed" json:"when_failed"`
			Flags      struct {
				Value         int    `yaml:"value" json:"value"`
				String        string `yaml:"string" json:"string"`
				Prefailure    bool   `yaml:"prefailure" json:"prefailure"`
				UpdatedOnline bool   `yaml:"updated_online" json:"updated_online"`
				Performance   bool   `yaml:"performance" json:"performance"`
				ErrorRate     bool   `yaml:"error_rate" json:"error_rate"`
				EventCount    bool   `yaml:"event_count" json:"event_count"`
				AutoKeep      bool   `yaml:"auto_keep" json:"auto_keep"`
			} `yaml:"flags" json:"flags"`
			Raw struct {
				Value  int    `yaml:"value" json:"value"`
				String string `yaml:"string" json:"string"`
			} `yaml:"raw" json:"raw"`
		} `yaml:"table" json:"table"`
	} `yaml:"ata_smart_attributes" json:"ata_smart_attributes"`
	PowerOnTime struct {
		Hours int `yaml:"hours" json:"hours"`
	} `yaml:"power_on_time" json:"power_on_time"`
	PowerCycleCount int `yaml:"power_cycle_count" json:"power_cycle_count"`
	Temperature     struct {
		Current int `yaml:"current" json:"current"`
	} `yaml:"temperature" json:"temperature"`
	AtaSmartErrorLog struct {
		Summary struct {
			Revision int `yaml:"revision" json:"revision"`
			Count    int `yaml:"count" json:"count"`
		} `yaml:"summary" json:"summary"`
	} `yaml:"ata_smart_error_log" json:"ata_smart_error_log"`
	AtaSmartSelfTestLog struct {
		Standard struct {
			Revision int `yaml:"revision" json:"revision"`
			Count    int `yaml:"count" json:"count"`
		} `yaml:"standard" json:"standard"`
	} `yaml:"ata_smart_self_test_log" json:"ata_smart_self_test_log"`
	AtaSmartSelectiveSelfTestLog struct {
		Revision int `yaml:"revision" json:"revision"`
		Table    []struct {
			LbaMin int `yaml:"lba_min" json:"lba_min"`
			LbaMax int `yaml:"lba_max" json:"lba_max"`
			Status struct {
				Value  int    `yaml:"value" json:"value"`
				String string `yaml:"string" json:"string"`
			} `yaml:"status" json:"status"`
		} `yaml:"table" json:"table"`
		Flags struct {
			Value                int  `yaml:"value" json:"value"`
			RemainderScanEnabled bool `yaml:"remainder_scan_enabled" json:"remainder_scan_enabled"`
		} `yaml:"flags" json:"flags"`
		PowerUpScanResumeMinutes int `yaml:"power_up_scan_resume_minutes" json:"power_up_scan_resume_minutes"`
	} `yaml:"ata_smart_selective_self_test_log" json:"ata_smart_selective_self_test_log"`
	Messages []struct {
		String   string `yaml:"string,omitempty" json:"string,omitempty"`
		Severity string `yaml:"severity,omitempty" json:"severity,omitempty"`
	} `yaml:"messages,omitempty" json:"messages,omitempty"`
}

// String implements stringer and returns a yaml formatted string
func (s *SMARTInfo) String() string {
	b, err := yaml.Marshal(s)
	if err != nil {
		panic(err)
	}

	return string(b)
}

// Healthy returns true if SMARTInfo.SmartStatus.Passed is true
func (s *SMARTInfo) Healthy() bool {
	return s.SmartStatus.Passed
}

// Error returns true if SMARTInfo.Messages has an error value
func (s *SMARTInfo) Error() error {
	if s.Smartctl.ExitStatus == 0 {
		return nil
	}

	var buf bytes.Buffer

	fmt.Fprintf(&buf, "smartctl error: %s", s.ExitCode)

	for _, msg := range s.Messages {
		fmt.Fprintf(&buf, " [%s] %s", msg.Severity, msg.String)
	}

	return errors.New(buf.String())
}

// JSON marshals SMARTInfo to json
func (s *SMARTInfo) JSON() ([]byte, error) {
	return json.Marshal(s)
}

// GetSMARTInfo creates a new SMARTInfo for given device
func GetSMARTInfo(dev string) (*SMARTInfo, error) {
	if !smartctlInstalled {
		return nil, ErrSmartctlNotInstalled
	}

	var out bytes.Buffer

	cmd := exec.Command("smartctl", "-a", dev, "--json")
	cmd.Stdout = &out
	err := cmd.Run()

	if err != nil && out.Len() == 0 {
		return nil, err
	}

	info := new(SMARTInfo)
	if err = json.Unmarshal(out.Bytes(), info); err != nil {
		return nil, err
	}

	info.ExitCode = SMARTExitCode(info.Smartctl.ExitStatus)

	return info, nil
}

// SmartctlInstalled returns true if installed
func SmartctlInstalled() bool {
	return smartctlInstalled
}

var smartctlInstalled bool

func init() {
	smartctlInstalled = exec.Command("smartctl", "-h").Run() == nil
}
