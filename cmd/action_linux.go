package cmd

import (
	"fmt"
	"time"

	"github.com/InazumaV/V2bX/common/exec"
	"github.com/spf13/cobra"
)

var (
	startCommand = cobra.Command{
		Use:   "start",
		Short: "Start V2bX service",
		Run:   startHandle,
	}
	stopCommand = cobra.Command{
		Use:   "stop",
		Short: "Stop V2bX service",
		Run:   stopHandle,
	}
	restartCommand = cobra.Command{
		Use:   "restart",
		Short: "Restart V2bX service",
		Run:   restartHandle,
	}
	logCommand = cobra.Command{
		Use:   "log",
		Short: "Output V2bX log",
		Run: func(_ *cobra.Command, _ []string) {
			exec.RunCommandStd("journalctl", "-u", "V2bX.service", "-e", "--no-pager", "-f")
		},
	}
)

func init() {
	command.AddCommand(&startCommand)
	command.AddCommand(&stopCommand)
	command.AddCommand(&restartCommand)
	command.AddCommand(&logCommand)
}

func startHandle(_ *cobra.Command, _ []string) {
	r, err := checkRunning()
	if err != nil {
		fmt.Println(Err("check status error: ", err))
		fmt.Println(Err("V2bX启动失败"))
		return
	}
	if r {
		fmt.Println(Ok("V2bX已运行，无需再次启动，如需重启请选择重启"))
	}
	_, err = exec.RunCommandByShell("systemctl start V2bX.service")
	if err != nil {
		fmt.Println(Err("exec start cmd error: ", err))
		fmt.Println(Err("V2bX启动失败"))
		return
	}
	time.Sleep(time.Second * 3)
	r, err = checkRunning()
	if err != nil {
		fmt.Println(Err("check status error: ", err))
		fmt.Println(Err("V2bX启动失败"))
	}
	if !r {
		fmt.Println(Err("V2bX可能启动失败，请稍后使用 V2bX log 查看日志信息"))
		return
	}
	fmt.Println(Ok("V2bX 启动成功，请使用 V2bX log 查看运行日志"))
}

func stopHandle(_ *cobra.Command, _ []string) {
	_, err := exec.RunCommandByShell("systemctl stop V2bX.service")
	if err != nil {
		fmt.Println(Err("exec stop cmd error: ", err))
		fmt.Println(Err("V2bX停止失败"))
		return
	}
	time.Sleep(2 * time.Second)
	r, err := checkRunning()
	if err != nil {
		fmt.Println(Err("check status error:", err))
		fmt.Println(Err("V2bX停止失败"))
		return
	}
	if r {
		fmt.Println(Err("V2bX停止失败，可能是因为停止时间超过了两秒，请稍后查看日志信息"))
		return
	}
	fmt.Println(Ok("V2bX 停止成功"))
}

func restartHandle(_ *cobra.Command, _ []string) {
	_, err := exec.RunCommandByShell("systemctl restart V2bX.service")
	if err != nil {
		fmt.Println(Err("exec restart cmd error: ", err))
		fmt.Println(Err("V2bX重启失败"))
		return
	}
	r, err := checkRunning()
	if err != nil {
		fmt.Println(Err("check status error: ", err))
		fmt.Println(Err("V2bX重启失败"))
		return
	}
	if !r {
		fmt.Println(Err("V2bX可能启动失败，请稍后使用 V2bX log 查看日志信息"))
		return
	}
	fmt.Println(Ok("V2bX重启成功"))
}
