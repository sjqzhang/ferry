package worker

import (
	"context"
	"ferry/pkg/logger"
	"github.com/RichardKnop/machinery/v1/tasks"
	"os/exec"
	"path"
	"syscall"
)

var asyncTaskMap map[string]interface{}

func executeTaskBase(scriptPath string, params string) (err error) {
	//取文件后缀 .py .sh .req
	suffix := path.Ext(scriptPath)
	if suffix != ".py" && suffix != ".sh" && suffix != ".req" {
		logger.Errorf("目前仅支持Python、Shell、Http脚本的执行，请知悉。")
		return
	}

	command := exec.Command("python",scriptPath, params) //初始化Cmd
	out, err := command.CombinedOutput()
	if err != nil {
		logger.Errorf("task exec failed，%v", err.Error())
		return
	}
	logger.Info("Output: ", string(out))
	logger.Info("ProcessState PID: ", command.ProcessState.Pid())
	logger.Info("Exit Code ", command.ProcessState.Sys().(syscall.WaitStatus).ExitStatus())
	return
}

// ExecCommand 异步任务
func ExecCommand(classify string, scriptPath string, params string) (err error) {

	_, err = factory.GetExecutor(classify).ExecuteTask(context.TODO(), scriptPath, params)

	return err

}

func SendTask(ctx context.Context, classify string, scriptPath string, params string) {
	args := make([]tasks.Arg, 0)
	args = append(args, tasks.Arg{
		Name:  "classify",
		Type:  "string",
		Value: classify,
	})
	args = append(args, tasks.Arg{
		Name:  "scriptPath",
		Type:  "string",
		Value: scriptPath,
	})
	args = append(args, tasks.Arg{
		Name:  "params",
		Type:  "string",
		Value: params,
	})
	task, _ := tasks.NewSignature("ExecCommandTask", args)
	task.RetryCount = 5
	_, err := AsyncTaskCenter.SendTaskWithContext(ctx, task)
	if err != nil {
		logger.Error(err.Error())
	}
}

func initAsyncTaskMap() {
	asyncTaskMap = make(map[string]interface{})
	asyncTaskMap["ExecCommandTask"] = ExecCommand
}
