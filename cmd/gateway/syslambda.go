package main

import (
	"../../proto"
	"fmt"
	"github.com/sirupsen/logrus"
	"os/exec"
)

func SysLambda(p pb.ThingMSG) {
	logrus.Info(BaseDir)
	switch p.Func {
	case "shell":
		logrus.Info("start sys shell")
		cmd := fmt.Sprintf("%s/bin/syslambda -app shell -url wss://www.icetears.com/v1/devices/device-id/console -cert %s/conf/certs/client.crt -key %s/conf/private_keys/client.key", BaseDir, BaseDir, BaseDir)

		o, err := exec.Command("bash", "-c", cmd).Output()
		if err != nil {
			logrus.Info(o, err)
		}
		logrus.Info("shell exit")
	}

}
