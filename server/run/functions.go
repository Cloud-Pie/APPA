package run

import (
	"strconv"
	"time"
	"strings"
	"fmt"
	"os/exec"
	"log"
)

func ValueAssignString(value *string, fallback string) string{

	if value!=nil {
		return *value
	} else {
		return fallback
	}
}
func ValueAssignInt64(value *int64, fallback int64) int64{

	if value!=nil {
		return *value
	} else {
		return fallback
	}
}

func StringInSlice(a string, list []string) bool {
	for _, b := range list {
		if b == a {
			return true
		}
	}
	return false
}
func FloatToString(input_num float64) string {
	// to convert a float number to a string
	return strconv.FormatFloat(input_num, 'f', 6, 64)
}
func StringToFloat(stringVal string) float64 {
	// to convert a float number to a string
	if s, err := strconv.ParseFloat(stringVal, 64); err == nil {
		return s// 3.14159265
	}
	return 0
}

func GetImageFromRegion(region string) string{
	for index, b := range DefaultRegion {
		if b == region {
			return DefaultAMI[index]
		}
	}
	return ""
}
func GetZoneFromRegion(region string) string{
	for index, b := range DefaultRegion {
		if b == region {
			return DefaultZone[index]
		}
	}
	return ""
}
func Schedule(what func(), delay time.Duration) chan bool {
	stop := make(chan bool)

	go func() {
		for {
			what()
			select {
			case <-time.After(delay):
			case <-stop:
				return
			}
		}
	}()

	return stop
}


func exe_cmd_output(cmd string) string {
	log.Println("Command : ",cmd)
	// splitting head => g++ parts => rest of the command
	parts := strings.Fields(cmd)
	head := parts[0]
	parts = parts[1:len(parts)]

	fmt.Println(parts)

	out, err := exec.Command(head,parts...).Output()
	if err != nil {
		fmt.Printf("%s", err)
		return ""
	}else{
		fmt.Printf("%s", out)
		return string(out)
	}

}