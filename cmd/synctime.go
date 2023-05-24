package cmd

import (
	"fmt"
	"github.com/beevik/ntp"
	"github.com/sagernet/sing-box/common/settings"
	"github.com/spf13/cobra"
)

var ntpServer string

var commandSyncTime = &cobra.Command{
	Use:   "synctime",
	Short: "Sync time from ntp server",
	Args:  cobra.NoArgs,
	Run:   synctimeHandle,
}

func init() {
	commandSyncTime.Flags().StringVar(&ntpServer, "server", "time.apple.com", "ntp server")
	command.AddCommand(commandSyncTime)
}

func synctimeHandle(_ *cobra.Command, _ []string) {
	t, err := ntp.Time(ntpServer)
	if err != nil {
		fmt.Println(Err("get time from server error: ", err))
		fmt.Println(Err("同步时间失败"))
		return
	}
	err = settings.SetSystemTime(t)
	if err != nil {
		fmt.Println(Err("set system time error: ", err))
		fmt.Println(Err("同步时间失败"))
		return
	}
	fmt.Println(Ok("同步时间成功"))
}
