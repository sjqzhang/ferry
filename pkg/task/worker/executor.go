package worker

import (
	"bytes"
	"context"
	"encoding/json"
	"ferry/pkg/logger"
	"fmt"
	"io/ioutil"
	"net/http"
	"os/exec"
	"syscall"
)

type IExecutor interface {
	ExecuteTask(ctx context.Context, scriptPath string, params string) (string, error)
}

type executorFactory struct {
	GetExecutor func(suffix string) IExecutor
}

var factory *executorFactory

func init() {
	factory = &executorFactory{
		GetExecutor: func(suffix string) IExecutor {
			if suffix == "python" {
				return &PythonExecutor{}
			} else if suffix == "shell" {
				return &ShellExecutor{}
			} else if suffix == "http" {
				return &HttpExecutor{}
			} else {
				logger.Errorf("目前仅支持Python、Shell、Http脚本的执行，请知悉。")
				return &DefaultExecutor{}
			}
		},
	}
}

type PythonExecutor struct {
}

type ShellExecutor struct {
}

type HttpExecutor struct {
}

type DefaultExecutor struct {
}

func (p *PythonExecutor) ExecuteTask(ctx context.Context, scriptPath string, params string) (string, error) {

	command := exec.CommandContext(ctx, "python", scriptPath, params) //初始化Cmd
	out, err := command.CombinedOutput()
	if err != nil {
		logger.Errorf("task exec failed，%v", err.Error())
		return "", err
	}
	logger.Info("Output: ", string(out))
	logger.Info("ProcessState PID: ", command.ProcessState.Pid())
	logger.Info("Exit Code ", command.ProcessState.Sys().(syscall.WaitStatus).ExitStatus())
	return string(out), nil
}

func (s *ShellExecutor) ExecuteTask(ctx context.Context, scriptPath string, params string) (string, error) {
	command := exec.CommandContext(ctx, "bash", scriptPath, params) //初始化Cmd
	out, err := command.CombinedOutput()
	if err != nil {
		logger.Errorf("task exec failed，%v", err.Error())
		return "", err
	}
	logger.Info("Output: ", string(out))
	logger.Info("ProcessState PID: ", command.ProcessState.Pid())
	logger.Info("Exit Code ", command.ProcessState.Sys().(syscall.WaitStatus).ExitStatus())
	return string(out), nil
}

func (h *HttpExecutor) ExecuteTask(ctx context.Context, scriptPath string, params string) (string, error) {
	//使用 http.Request 发送Post请求
	var data map[string]interface{}
	err := json.Unmarshal([]byte(params), &data)
	logger.Error(err)
	sendData:=make(map[string]interface{})
	sendData["data"] = data
	reqBody, err := json.Marshal(sendData)
	if err != nil {
		logger.Errorf("task exec failed，%v", err.Error())
		return "", err
	}
	req,err:=http.NewRequest("POST",scriptPath,bytes.NewBuffer(reqBody))
	req.Header.Set("Content-Type","application/json")

	//发送请求
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		logger.Errorf("task exec failed，%v", err.Error())
		return "", err
	}
	defer resp.Body.Close()
	respBody, err:= ioutil.ReadAll(resp.Body)
	if err != nil {
		logger.Errorf("task exec failed，%v", err.Error())
		return "", err
	}
	logger.Info("Output: ", string(respBody))
	return string(respBody), nil

}

func (d *DefaultExecutor) ExecuteTask(ctx context.Context, scriptPath string, params string) (string, error) {
	return "", fmt.Errorf("未实现")
}
