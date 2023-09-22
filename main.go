package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"os/exec"
	"path/filepath"
	"regexp"

	rpc "github.com/disaipe/dev01-rpc-base"
)

type ComputerStatusItem struct {
	Id       int
	UserName string
}

type GetComputerStateRequest struct {
	rpc.Response

	Id       int
	Parallel bool
	Host     string
	Hosts    []GetComputerStateRequest
}

type GetComputerStateResponse struct {
	rpc.ResultResponse

	Id     int
	Status bool
	Error  string
	Hosts  []ComputerStatusItem
}

func GetComputerState(hosts []string) ([]ComputerStatusItem, error) {
	var computerStatusItems []ComputerStatusItem

	scriptPath := filepath.Join(rpc.Config.GetWorkingDir(), "script.ps1")

	args := append([]string{
		"-nologo",
		"-noprofile",
		"-NonInteractive",
		"-ExecutionPolicy", "ByPass",
		"-OutputFormat", "Text",
		"-File", scriptPath,
	}, hosts...)

	cmd := exec.Command("powershell.exe", args...)
	cmd.Dir = rpc.Config.GetWorkingDir()

	out, err := cmd.CombinedOutput()
	monitor
		if err != nil {
			return nil, err
		}
	}

	return computerStatusItems, nil
}

var rpcAction = rpc.ActionFunction(func(rpcServer *rpc.Rpc, body io.ReadCloser, appAuth string) (rpc.Response, error) {
	var computerStateRequest GetComputerStateRequest

	err := json.NewDecoder(body).Decode(&computerStateRequest)

	if err != nil {
		return nil, err
	}

	var hosts = []string{}
	var resultStatus = true
	var resultMessage string

	if computerStateRequest.Hosts != nil {
		for _, computer := range computerStateRequest.Hosts {
			if computer.Id != 0 && computer.Host != "" {
				hosts = append(hosts, fmt.Sprintf("%d:%s", computer.Id, computer.Host))
			}
		}
	} else {
		if computerStateRequest.Id == 0 {
			resultStatus = false
			resultMessage = "Id is required"
		} else if computerStateRequest.Host == "" {
			resultStatus = false
			resultMessage = "Host is required"
		} else {
			hosts = append(hosts, fmt.Sprintf("%d:%s", computerStateRequest.Id, computerStateRequest.Host))
		}
	}

	if len(hosts) != 0 {
		var action = func() error {
			items, err := GetComputerState(hosts)

			resultData := &GetComputerStateResponse{
				Id:     computerStateRequest.Id,
				Status: err == nil,
				Hosts:  items,
			}

			if err != nil {
				return err
			}

			go func() {
				rpcServer.SendResult(*resultData, appAuth)
			}()

			return nil
		}

		go func() {
			if computerStateRequest.Parallel {
				action()
			} else {
				rpcServer.AddJob(rpc.Job{Action: action})
			}
		}()
	}

	requestAcceptedResponse := &rpc.ActionResponse{
		Status: resultStatus,
		Data:   resultMessage,
	}

	return requestAcceptedResponse, nil
})

func main() {
	flag.Parse()

	rpc.Config.SetServiceSettings(
		"dev01-pcmon-daemon",
		"Dev01 Computer state monitor",
		"The part of the Dev01 platform",
	)

	rpc.Config.SetAction("/computer_sync_job", &rpcAction)

	if rpc.Config.Serving() {
		rpcServer := &rpc.Rpc{}
		rpcServer.Run()
	}
}
