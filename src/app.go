package main

import (
	"encoding/xml"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"os/exec"
	"regexp"
	"strconv"
)

const LISTEN_ADDRESS = ":9202"
const NVIDIA_SMI_PATH = "/usr/bin/nvidia-smi"

var testMode string

type NvidiaSmiLog struct {
	DriverVersion string `xml:"driver_version"`
	CudaVersion   string `xml:"cuda_version"`
	AttachedGPUs  string `xml:"attached_gpus"`
	GPU           []struct {
		Id                       string `xml:"id,attr"`
		ProductName              string `xml:"product_name"`
		ProductBrand             string `xml:"product_brand"`
		DisplayMode              string `xml:"display_mode"`
		DisplayActive            string `xml:"display_active"`
		PersistenceMode          string `xml:"persistence_mode"`
		AccountingMode           string `xml:"accounting_mode"`
		AccountingModeBufferSize string `xml:"accounting_mode_buffer_size"`
		DriverModel              struct {
			CurrentDM string `xml:"current_dm"`
			PendingDM string `xml:"pending_dm"`
		} `xml:"driver_model"`
		Serial         string `xml:"serial"`
		UUID           string `xml:"uuid"`
		MinorNumber    string `xml:"minor_number"`
		VbiosVersion   string `xml:"vbios_version"`
		MultiGPUBoard  string `xml:"multigpu_board"`
		BoardId        string `xml:"board_id"`
		GPUPartNumber  string `xml:"gpu_part_number"`
		InfoRomVersion struct {
			ImgVersion string `xml:"img_version"`
			OemObject  string `xml:"oem_object"`
			EccObject  string `xml:"ecc_object"`
			PwrObject  string `xml:"pwr_object"`
		} `xml:"inforom_version"`
		GPUOperationMode struct {
			Current string `xml:"current_gom"`
			Pending string `xml:"pending_gom"`
		} `xml:"gpu_operation_mode"`
		GPUVirtualizationMode struct {
			VirtualizationMode string `xml:"virtualization_mode"`
			HostVGPUMode       string `xml:"host_vgpu_mode"`
		} `xml:"gpu_virtualization_mode"`
		IBMNPU struct {
			RelaxedOrderingMode string `xml:"relaxed_ordering_mode"`
		} `xml:"ibmnpu"`
		PCI struct {
			Bus         string `xml:"pci_bus"`
			Device      string `xml:"pci_device"`
			Domain      string `xml:"pci_domain"`
			DeviceId    string `xml:"pci_device_id"`
			BusId       string `xml:"pci_bus_id"`
			SubSystemId string `xml:"pci_sub_system_id"`
			GPULinkInfo struct {
				PCIeGen struct {
					Max     string `xml:"max_link_gen"`
					Current string `xml:"current_link_gen"`
				} `xml:"pcie_gen"`
				LinkWidth struct {
					Max     string `xml:"max_link_width"`
					Current string `xml:"current_link_width"`
				} `xml:"link_widths"`
			} `xml:"pci_gpu_link_info"`
			BridgeChip struct {
				Type string `xml:"bridge_chip_type"`
				Fw   string `xml:"bridge_chip_fw"`
			} `xml:"pci_bridge_chip"`
			ReplayCounter         string `xml:"replay_counter"`
			ReplayRolloverCounter string `xml:"replay_rollover_counter"`
			TxUtil                string `xml:"tx_util"`
			RxUtil                string `xml:"rx_util"`
		} `xml:"pci"`
		FanSpeed         string `xml:"fan_speed"`
		PerformanceState string `xml:"performance_state"`
		// <clocks_throttle_reasons>
		// 	    <clocks_throttle_reason_gpu_idle>Not Active</clocks_throttle_reason_gpu_idle>
		// 	    <clocks_throttle_reason_applications_clocks_setting>Not Active</clocks_throttle_reason_applications_clocks_setting>
		// 	    <clocks_throttle_reason_sw_power_cap>Not Active</clocks_throttle_reason_sw_power_cap>
		// 	    <clocks_throttle_reason_hw_slowdown>Not Active</clocks_throttle_reason_hw_slowdown>
		// 	    <clocks_throttle_reason_hw_thermal_slowdown>N/A</clocks_throttle_reason_hw_thermal_slowdown>
		// 	    <clocks_throttle_reason_hw_power_brake_slowdown>N/A</clocks_throttle_reason_hw_power_brake_slowdown>
		// 	    <clocks_throttle_reason_sync_boost>Not Active</clocks_throttle_reason_sync_boost>
		// 	    <clocks_throttle_reason_sw_thermal_slowdown>Not Active</clocks_throttle_reason_sw_thermal_slowdown>
		// 	    <clocks_throttle_reason_display_clocks_setting>Not Active</clocks_throttle_reason_display_clocks_setting>
		// </clocks_throttle_reasons>
		FbMemoryUsage struct {
			Total string `xml:"total"`
			Used  string `xml:"used"`
			Free  string `xml:"free"`
		} `xml:"fb_memory_usage"`
		Bar1MemoryUsage struct {
			Total string `xml:"total"`
			Used  string `xml:"used"`
			Free  string `xml:"free"`
		} `xml:"bar1_memory_usage"`
		ComputeMode string `xml:"compute_mode"`
		Utilization struct {
			GPUUtil     string `xml:"gpu_util"`
			MemoryUtil  string `xml:"memory_util"`
			EncoderUtil string `xml:"encoder_util"`
			DecoderUtil string `xml:"decoder_util"`
		} `xml:"utilization"`
		EncoderStats struct {
			SessionCount   string `xml:"session_count"`
			AverageFPS     string `xml:"average_fps"`
			AverageLatency string `xml:"average_latency"`
		} `xml:"encoder_stats"`
		FBCStats struct {
			SessionCount   string `xml:"session_count"`
			AverageFPS     string `xml:"average_fps"`
			AverageLatency string `xml:"average_latency"`
		} `xml:"fbc_stats"`
		// <ecc_mode>
		//     <current_ecc>N/A</current_ecc>
		//     <pending_ecc>N/A</pending_ecc>
		// </ecc_mode>
		// <ecc_errors>
		//     <volatile>
		//         <single_bit>
		//             <device_memory>N/A</device_memory>
		//             <register_file>N/A</register_file>
		//             <l1_cache>N/A</l1_cache>
		//             <l2_cache>N/A</l2_cache>
		//             <texture_memory>N/A</texture_memory>
		//             <texture_shm>N/A</texture_shm>
		//             <cbu>N/A</cbu>
		//             <total>N/A</total>
		//         </single_bit>
		//         <double_bit>
		//             <device_memory>N/A</device_memory>
		//             <register_file>N/A</register_file>
		//             <l1_cache>N/A</l1_cache>
		//             <l2_cache>N/A</l2_cache>
		//             <texture_memory>N/A</texture_memory>
		//             <texture_shm>N/A</texture_shm>
		//             <cbu>N/A</cbu>
		//             <total>N/A</total>
		//         </double_bit>
		//     </volatile>
		//     <aggregate>
		//         <single_bit>
		//             <device_memory>N/A</device_memory>
		//             <register_file>N/A</register_file>
		//             <l1_cache>N/A</l1_cache>
		//             <l2_cache>N/A</l2_cache>
		//             <texture_memory>N/A</texture_memory>
		//             <texture_shm>N/A</texture_shm>
		//             <cbu>N/A</cbu>
		//             <total>N/A</total>
		//         </single_bit>
		//         <double_bit>
		//             <device_memory>N/A</device_memory>
		//             <register_file>N/A</register_file>
		//             <l1_cache>N/A</l1_cache>
		//             <l2_cache>N/A</l2_cache>
		//             <texture_memory>N/A</texture_memory>
		//             <texture_shm>N/A</texture_shm>
		//             <cbu>N/A</cbu>
		//             <total>N/A</total>
		//         </double_bit>
		//     </aggregate>
		// </ecc_errors>
		// <retired_pages>
		//     <multiple_single_bit_retirement>
		//         <retired_count>N/A</retired_count>
		//         <retired_pagelist>N/A</retired_pagelist>
		//     </multiple_single_bit_retirement>
		//     <double_bit_retirement>
		//         <retired_count>N/A</retired_count>
		//         <retired_pagelist>N/A</retired_pagelist>
		//     </double_bit_retirement>
		//     <pending_blacklist>N/A</pending_blacklist>
		//     <pending_retirement>N/A</pending_retirement>
		// </retired_pages>
		Temperature struct {
			GPUTemp                string `xml:"gpu_temp"`
			GPUTempMaxThreshold    string `xml:"gpu_temp_max_threshold"`
			GPUTempSlowThreshold   string `xml:"gpu_temp_slow_threshold"`
			GPUTempMaxGpuThreshold string `xml:"gpu_temp_max_gpu_threshold"`
			MemoryTemp             string `xml:"memory_temp"`
			GPUTempMaxMemThreshold string `xml:"gpu_temp_max_mem_threshold"`
		} `xml:"temperature"`
		PowerReadings struct {
			PowerState         string `xml:"power_state"`
			PowerDraw          string `xml:"power_draw"`
			PowerLimit         string `xml:"power_limit"`
			DefaultPowerLimit  string `xml:"default_power_limit"`
			EnforcedPowerLimit string `xml:"enforced_power_limit"`
			MinPowerLimit      string `xml:"min_power_limit"`
			MaxPowerLimit      string `xml:"max_power_limit"`
		} `xml:"power_readings gpu_power_readings"`
		Clocks struct {
			GraphicsClock string `xml:"graphics_clock"`
			SmClock       string `xml:"sm_clock"`
			MemClock      string `xml:"mem_clock"`
			VideoClock    string `xml:"video_clock"`
		} `xml:"clocks"`
		// <applications_clocks>
		// 	<graphics_clock>1190 MHz</graphics_clock>
		// 	<mem_clock>3505 MHz</mem_clock>
		// </applications_clocks>
		// <default_applications_clocks>
		// 	<graphics_clock>1190 MHz</graphics_clock>
		// 	<mem_clock>3505 MHz</mem_clock>
		// </default_applications_clocks>
		MaxClocks struct {
			GraphicsClock string `xml:"graphics_clock"`
			SmClock       string `xml:"sm_clock"`
			MemClock      string `xml:"mem_clock"`
			VideoClock    string `xml:"video_clock"`
		} `xml:"max_clocks"`
		// <max_customer_boost_clocks>
		// 	<graphics_clock>N/A</graphics_clock>
		// </max_customer_boost_clocks>
		ClockPolicy struct {
			AutoBoost        string `xml:"auto_boost"`
			AutoBoostDefault string `xml:"auto_boost_default"`
		} `xml:"clock_policy"`
		// <supported_clocks>
		//     <supported_mem_clock>
		//         [...]
		//     </supported_mem_clock>
		// </supported_clocks>
		Processes struct {
			ProcessInfo []struct {
				Pid         string `xml:"pid"`
				Type        string `xml:"type"`
				ProcessName string `xml:"process_name"`
				UsedMemory  string `xml:"used_memory"`
			} `xml:"process_info"`
		} `xml:"processes"`
		// <accounted_processes>
		// </accounted_processes>
	} `xml:"gpu"`
}

func formatVersion(key string, meta string, value string) string {
	r := regexp.MustCompile(`(?P<version>\d+\.\d+).*`)
	match := r.FindStringSubmatch(value)
	version := "0"
	if len(match) > 0 {
		version = match[1]
	}
	return formatValue(key, meta, version)
}

func formatValue(key string, meta string, value string) string {
	result := key
	if meta != "" {
		result += "{" + meta + "}"
	}
	return result + " " + value + "\n"
}

func filterUnit(s string) string {
	if s == "N/A" {
		return "0"
	}
	r := regexp.MustCompile(`(?P<value>[\d\.]+) (?P<power>[KMGT]?[i]?)(?P<unit>.*)`)
	match := r.FindStringSubmatch(s)
	if len(match) == 0 {
		return "0"
	}

	result := make(map[string]string)
	for i, name := range r.SubexpNames() {
		if i != 0 && name != "" {
			result[name] = match[i]
		}
	}

	power := result["power"]
	if value, err := strconv.ParseFloat(result["value"], 32); err == nil {
		switch power {
		case "K":
			value *= 1000
		case "M":
			value *= 1000 * 1000
		case "G":
			value *= 1000 * 1000 * 1000
		case "T":
			value *= 1000 * 1000 * 1000 * 1000
		case "Ki":
			value *= 1024
		case "Mi":
			value *= 1024 * 1024
		case "Gi":
			value *= 1024 * 1024 * 1024
		case "Ti":
			value *= 1024 * 1024 * 1024 * 1024
		}
		return fmt.Sprintf("%g", value)
	}
	return "0"
}

func filterNumber(value string) string {
	if value == "N/A" {
		return "0"
	}
	r := regexp.MustCompile("[^0-9.]")
	return r.ReplaceAllString(value, "")
}

func metrics(w http.ResponseWriter, r *http.Request) {
	log.Print("Serving /metrics")

	var cmd *exec.Cmd
	if testMode == "1" {
		dir, err := os.Getwd()
		if err != nil {
			log.Fatal(err)
		}
		cmd = exec.Command("/bin/cat", dir+"/nvidia-smi.sample.xml")
	} else {
		cmd = exec.Command(NVIDIA_SMI_PATH, "-q", "-x")
	}

	// Execute system command
	stdout, err := cmd.Output()
	if err != nil {
		println(err.Error())
		if testMode != "1" {
			println("Something went wrong with the execution of nvidia-smi")
		}
		return
	}

	// Parse XML
	var xmlData NvidiaSmiLog
	xml.Unmarshal(stdout, &xmlData)

	// Output
	for _, GPU := range xmlData.GPU {
		io.WriteString(w, formatVersion("nvidiasmi_driver_version", "id=\""+GPU.Id+"\",uuid=\""+GPU.UUID+"\",name=\""+GPU.ProductName+"\"", xmlData.DriverVersion))
		io.WriteString(w, formatVersion("nvidiasmi_cuda_version", "id=\""+GPU.Id+"\",uuid=\""+GPU.UUID+"\",name=\""+GPU.ProductName+"\"", xmlData.CudaVersion))
		io.WriteString(w, formatValue("nvidiasmi_attached_gpus", "id=\""+GPU.Id+"\",uuid=\""+GPU.UUID+"\",name=\""+GPU.ProductName+"\"", xmlData.AttachedGPUs))
		io.WriteString(w, formatValue("nvidiasmi_pci_pcie_gen_max", "id=\""+GPU.Id+"\",uuid=\""+GPU.UUID+"\",name=\""+GPU.ProductName+"\"", GPU.PCI.GPULinkInfo.PCIeGen.Max))
		io.WriteString(w, formatValue("nvidiasmi_pci_pcie_gen_current", "id=\""+GPU.Id+"\",uuid=\""+GPU.UUID+"\",name=\""+GPU.ProductName+"\"", GPU.PCI.GPULinkInfo.PCIeGen.Current))
		io.WriteString(w, formatValue("nvidiasmi_pci_link_width_max_multiplicator", "id=\""+GPU.Id+"\",uuid=\""+GPU.UUID+"\",name=\""+GPU.ProductName+"\"", filterNumber(GPU.PCI.GPULinkInfo.LinkWidth.Max)))
		io.WriteString(w, formatValue("nvidiasmi_pci_link_width_current_multiplicator", "id=\""+GPU.Id+"\",uuid=\""+GPU.UUID+"\",name=\""+GPU.ProductName+"\"", filterNumber(GPU.PCI.GPULinkInfo.LinkWidth.Current)))
		io.WriteString(w, formatValue("nvidiasmi_pci_replay_counter", "id=\""+GPU.Id+"\",uuid=\""+GPU.UUID+"\",name=\""+GPU.ProductName+"\"", GPU.PCI.ReplayRolloverCounter))
		io.WriteString(w, formatValue("nvidiasmi_pci_replay_rollover_counter", "id=\""+GPU.Id+"\",uuid=\""+GPU.UUID+"\",name=\""+GPU.ProductName+"\"", GPU.PCI.ReplayRolloverCounter))
		io.WriteString(w, formatValue("nvidiasmi_pci_tx_util_bytes_per_second", "id=\""+GPU.Id+"\",uuid=\""+GPU.UUID+"\",name=\""+GPU.ProductName+"\"", filterUnit(GPU.PCI.TxUtil)))
		io.WriteString(w, formatValue("nvidiasmi_pci_rx_util_bytes_per_second", "id=\""+GPU.Id+"\",uuid=\""+GPU.UUID+"\",name=\""+GPU.ProductName+"\"", filterUnit(GPU.PCI.RxUtil)))
		io.WriteString(w, formatValue("nvidiasmi_fan_speed_percent", "id=\""+GPU.Id+"\",uuid=\""+GPU.UUID+"\",name=\""+GPU.ProductName+"\"", filterUnit(GPU.FanSpeed)))
		io.WriteString(w, formatValue("nvidiasmi_performance_state_int", "id=\""+GPU.Id+"\",uuid=\""+GPU.UUID+"\",name=\""+GPU.ProductName+"\"", filterNumber(GPU.PerformanceState)))
		io.WriteString(w, formatValue("nvidiasmi_fb_memory_usage_total_bytes", "id=\""+GPU.Id+"\",uuid=\""+GPU.UUID+"\",name=\""+GPU.ProductName+"\"", filterUnit(GPU.FbMemoryUsage.Total)))
		io.WriteString(w, formatValue("nvidiasmi_fb_memory_usage_used_bytes", "id=\""+GPU.Id+"\",uuid=\""+GPU.UUID+"\",name=\""+GPU.ProductName+"\"", filterUnit(GPU.FbMemoryUsage.Used)))
		io.WriteString(w, formatValue("nvidiasmi_fb_memory_usage_free_bytes", "id=\""+GPU.Id+"\",uuid=\""+GPU.UUID+"\",name=\""+GPU.ProductName+"\"", filterUnit(GPU.FbMemoryUsage.Free)))
		io.WriteString(w, formatValue("nvidiasmi_bar1_memory_usage_total_bytes", "id=\""+GPU.Id+"\",uuid=\""+GPU.UUID+"\",name=\""+GPU.ProductName+"\"", filterUnit(GPU.Bar1MemoryUsage.Total)))
		io.WriteString(w, formatValue("nvidiasmi_bar1_memory_usage_used_bytes", "id=\""+GPU.Id+"\",uuid=\""+GPU.UUID+"\",name=\""+GPU.ProductName+"\"", filterUnit(GPU.Bar1MemoryUsage.Used)))
		io.WriteString(w, formatValue("nvidiasmi_bar1_memory_usage_free_bytes", "id=\""+GPU.Id+"\",uuid=\""+GPU.UUID+"\",name=\""+GPU.ProductName+"\"", filterUnit(GPU.Bar1MemoryUsage.Free)))
		io.WriteString(w, formatValue("nvidiasmi_utilization_gpu_percent", "id=\""+GPU.Id+"\",uuid=\""+GPU.UUID+"\",name=\""+GPU.ProductName+"\"", filterUnit(GPU.Utilization.GPUUtil)))
		io.WriteString(w, formatValue("nvidiasmi_utilization_memory_percent", "id=\""+GPU.Id+"\",uuid=\""+GPU.UUID+"\",name=\""+GPU.ProductName+"\"", filterUnit(GPU.Utilization.MemoryUtil)))
		io.WriteString(w, formatValue("nvidiasmi_utilization_encoder_percent", "id=\""+GPU.Id+"\",uuid=\""+GPU.UUID+"\",name=\""+GPU.ProductName+"\"", filterUnit(GPU.Utilization.EncoderUtil)))
		io.WriteString(w, formatValue("nvidiasmi_utilization_decoder_percent", "id=\""+GPU.Id+"\",uuid=\""+GPU.UUID+"\",name=\""+GPU.ProductName+"\"", filterUnit(GPU.Utilization.DecoderUtil)))
		io.WriteString(w, formatValue("nvidiasmi_encoder_session_count", "id=\""+GPU.Id+"\",uuid=\""+GPU.UUID+"\",name=\""+GPU.ProductName+"\"", GPU.EncoderStats.SessionCount))
		io.WriteString(w, formatValue("nvidiasmi_encoder_average_fps", "id=\""+GPU.Id+"\",uuid=\""+GPU.UUID+"\",name=\""+GPU.ProductName+"\"", GPU.EncoderStats.AverageFPS))
		io.WriteString(w, formatValue("nvidiasmi_encoder_average_latency", "id=\""+GPU.Id+"\",uuid=\""+GPU.UUID+"\",name=\""+GPU.ProductName+"\"", GPU.EncoderStats.AverageLatency))
		io.WriteString(w, formatValue("nvidiasmi_fbc_session_count", "id=\""+GPU.Id+"\",uuid=\""+GPU.UUID+"\",name=\""+GPU.ProductName+"\"", GPU.FBCStats.SessionCount))
		io.WriteString(w, formatValue("nvidiasmi_fbc_average_fps", "id=\""+GPU.Id+"\",uuid=\""+GPU.UUID+"\",name=\""+GPU.ProductName+"\"", GPU.FBCStats.AverageFPS))
		io.WriteString(w, formatValue("nvidiasmi_fbc_average_latency", "id=\""+GPU.Id+"\",uuid=\""+GPU.UUID+"\",name=\""+GPU.ProductName+"\"", GPU.FBCStats.AverageLatency))
		io.WriteString(w, formatValue("nvidiasmi_gpu_temp_celsius", "id=\""+GPU.Id+"\",uuid=\""+GPU.UUID+"\",name=\""+GPU.ProductName+"\"", filterUnit(GPU.Temperature.GPUTemp)))
		io.WriteString(w, formatValue("nvidiasmi_gpu_temp_max_threshold_celsius", "id=\""+GPU.Id+"\",uuid=\""+GPU.UUID+"\",name=\""+GPU.ProductName+"\"", filterUnit(GPU.Temperature.GPUTempMaxThreshold)))
		io.WriteString(w, formatValue("nvidiasmi_gpu_temp_slow_threshold_celsius", "id=\""+GPU.Id+"\",uuid=\""+GPU.UUID+"\",name=\""+GPU.ProductName+"\"", filterUnit(GPU.Temperature.GPUTempSlowThreshold)))
		io.WriteString(w, formatValue("nvidiasmi_gpu_temp_max_gpu_threshold_celsius", "id=\""+GPU.Id+"\",uuid=\""+GPU.UUID+"\",name=\""+GPU.ProductName+"\"", filterUnit(GPU.Temperature.GPUTempMaxGpuThreshold)))
		io.WriteString(w, formatValue("nvidiasmi_memory_temp_celsius", "id=\""+GPU.Id+"\",uuid=\""+GPU.UUID+"\",name=\""+GPU.ProductName+"\"", filterUnit(GPU.Temperature.MemoryTemp)))
		io.WriteString(w, formatValue("nvidiasmi_gpu_temp_max_mem_threshold_celsius", "id=\""+GPU.Id+"\",uuid=\""+GPU.UUID+"\",name=\""+GPU.ProductName+"\"", filterUnit(GPU.Temperature.GPUTempMaxMemThreshold)))
		io.WriteString(w, formatValue("nvidiasmi_power_state_int", "id=\""+GPU.Id+"\",uuid=\""+GPU.UUID+"\",name=\""+GPU.ProductName+"\"", filterNumber(GPU.PowerReadings.PowerState)))
		io.WriteString(w, formatValue("nvidiasmi_power_draw_watts", "id=\""+GPU.Id+"\",uuid=\""+GPU.UUID+"\",name=\""+GPU.ProductName+"\"", filterUnit(GPU.PowerReadings.PowerDraw)))
		io.WriteString(w, formatValue("nvidiasmi_power_limit_watts", "id=\""+GPU.Id+"\",uuid=\""+GPU.UUID+"\",name=\""+GPU.ProductName+"\"", filterUnit(GPU.PowerReadings.PowerLimit)))
		io.WriteString(w, formatValue("nvidiasmi_default_power_limit_watts", "id=\""+GPU.Id+"\",uuid=\""+GPU.UUID+"\",name=\""+GPU.ProductName+"\"", filterUnit(GPU.PowerReadings.DefaultPowerLimit)))
		io.WriteString(w, formatValue("nvidiasmi_enforced_power_limit_watts", "id=\""+GPU.Id+"\",uuid=\""+GPU.UUID+"\",name=\""+GPU.ProductName+"\"", filterUnit(GPU.PowerReadings.EnforcedPowerLimit)))
		io.WriteString(w, formatValue("nvidiasmi_min_power_limit_watts", "id=\""+GPU.Id+"\",uuid=\""+GPU.UUID+"\",name=\""+GPU.ProductName+"\"", filterUnit(GPU.PowerReadings.MinPowerLimit)))
		io.WriteString(w, formatValue("nvidiasmi_max_power_limit_watts", "id=\""+GPU.Id+"\",uuid=\""+GPU.UUID+"\",name=\""+GPU.ProductName+"\"", filterUnit(GPU.PowerReadings.MaxPowerLimit)))
		io.WriteString(w, formatValue("nvidiasmi_clock_graphics_hertz", "id=\""+GPU.Id+"\",uuid=\""+GPU.UUID+"\",name=\""+GPU.ProductName+"\"", filterUnit(GPU.Clocks.GraphicsClock)))
		io.WriteString(w, formatValue("nvidiasmi_clock_graphics_max_hertz", "id=\""+GPU.Id+"\",uuid=\""+GPU.UUID+"\",name=\""+GPU.ProductName+"\"", filterUnit(GPU.MaxClocks.GraphicsClock)))
		io.WriteString(w, formatValue("nvidiasmi_clock_sm_hertz", "id=\""+GPU.Id+"\",uuid=\""+GPU.UUID+"\",name=\""+GPU.ProductName+"\"", filterUnit(GPU.Clocks.SmClock)))
		io.WriteString(w, formatValue("nvidiasmi_clock_sm_max_hertz", "id=\""+GPU.Id+"\",uuid=\""+GPU.UUID+"\",name=\""+GPU.ProductName+"\"", filterUnit(GPU.MaxClocks.SmClock)))
		io.WriteString(w, formatValue("nvidiasmi_clock_mem_hertz", "id=\""+GPU.Id+"\",uuid=\""+GPU.UUID+"\",name=\""+GPU.ProductName+"\"", filterUnit(GPU.Clocks.MemClock)))
		io.WriteString(w, formatValue("nvidiasmi_clock_mem_max_hertz", "id=\""+GPU.Id+"\",uuid=\""+GPU.UUID+"\",name=\""+GPU.ProductName+"\"", filterUnit(GPU.MaxClocks.MemClock)))
		io.WriteString(w, formatValue("nvidiasmi_clock_video_hertz", "id=\""+GPU.Id+"\",uuid=\""+GPU.UUID+"\",name=\""+GPU.ProductName+"\"", filterUnit(GPU.Clocks.VideoClock)))
		io.WriteString(w, formatValue("nvidiasmi_clock_video_max_hertz", "id=\""+GPU.Id+"\",uuid=\""+GPU.UUID+"\",name=\""+GPU.ProductName+"\"", filterUnit(GPU.MaxClocks.VideoClock)))
		io.WriteString(w, formatValue("nvidiasmi_clock_policy_auto_boost", "id=\""+GPU.Id+"\",uuid=\""+GPU.UUID+"\",name=\""+GPU.ProductName+"\"", filterUnit(GPU.ClockPolicy.AutoBoost)))
		io.WriteString(w, formatValue("nvidiasmi_clock_policy_auto_boost_default", "id=\""+GPU.Id+"\",uuid=\""+GPU.UUID+"\",name=\""+GPU.ProductName+"\"", filterUnit(GPU.ClockPolicy.AutoBoostDefault)))
		for _, Process := range GPU.Processes.ProcessInfo {
			io.WriteString(w, formatValue("nvidiasmi_process_used_memory_bytes", "id=\""+GPU.Id+"\",uuid=\""+GPU.UUID+"\",name=\""+GPU.ProductName+"\",process_pid=\""+Process.Pid+"\",process_type=\""+Process.Type+"\"", filterUnit(Process.UsedMemory)))
		}
	}
}

func index(w http.ResponseWriter, r *http.Request) {
	log.Print("Serving /index")
	html := `<!doctype html>
<html>
    <head>
        <meta charset="utf-8">
        <title>Nvidia SMI Exporter</title>
    </head>
    <body>
        <h1>Nvidia SMI Exporter</h1>
        <p><a href="/metrics">Metrics</a></p>
    </body>
</html>`
	io.WriteString(w, html)
}

func main() {
	testMode = os.Getenv("TEST_MODE")
	if testMode == "1" {
		log.Print("Test mode is enabled")
	}

	log.Print("Nvidia SMI exporter listening on " + LISTEN_ADDRESS)
	http.HandleFunc("/", index)
	http.HandleFunc("/metrics", metrics)
	http.ListenAndServe(LISTEN_ADDRESS, nil)
}
