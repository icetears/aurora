package main

import (
	"../../proto"
	"github.com/gin-gonic/gin"
	"github.com/golang/protobuf/proto"
	//"github.com/jinzhu/gorm"
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"github.com/google/uuid"
	"io/ioutil"
)

type Device struct {
	ID      uuid.UUID
	UserID  int64
	Name    string
	ModelID string
	SN      string
	Status  int32
	Key     string
}

func DeviceCreateHandler(c *gin.Context) {
	var evt struct {
		Name    string `json:"name"`
		ModelID string `json:"modelID"`
		SN      string `json:"sn"`
		Key     string `json:"key"`
	}
	var dev Device
	buf, _ := ioutil.ReadAll(c.Request.Body)
	u := uuid.NewHash(sha256.New(), uuid.NameSpaceDNS, []byte(evt.SN), 5)
	json.Unmarshal(buf, &evt)
	dev.ID = u
	dev.Name = evt.Name
	dev.ModelID = evt.ModelID
	dev.SN = evt.SN
	dev.Key = evt.Key
	db.Save(&dev)
}

func DeviceDeleteHandler(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(404, gin.H{
			"message": err.Error(),
		})
	}
	if err := db.Where("id = ?", id.String()).Unscoped().Delete(&Device{}).Error; err != nil {
		c.JSON(404, gin.H{
			"message": err.Error(),
		})
	}
	c.JSON(200, gin.H{})
}

func DeviceListHandler(c *gin.Context) {
	var devices []Device
	db.Model(Device{}).Scan(&devices).Limit(20)
	c.JSON(200, devices)
}
func DeviceShadowHandler(c *gin.Context) {
}

func DeviceShadowUpdateHandler(c *gin.Context) {
}

func AppDeployHandler(c *gin.Context) {
	var (
		evt = struct {
			Name    string   `json:"name"`
			Version string   `json:"version"`
			Cmd     []string `json:"cmd"`
			Args    []string `json:"args"`
			Env     []string `json:"env"`
		}{}
		err error
		buf []byte
	)
	buf, err = ioutil.ReadAll(c.Request.Body)
	if err == nil {
		json.Unmarshal(buf, &evt)
		p := pb.ThingMSG{
			MsgId:   111,
			Cid:     "test",
			MsgType: pb.ThingMSG_DOCKER,
			Func:    fmt.Sprintf("%s:%s", evt.Name, evt.Version),
			Cmd:     []string{},
			Args:    evt.Args,
			Env:     evt.Env,
		}
		d, _ := proto.Marshal(&p)
		SendMSG(fmt.Sprintf("system/devices/%s", c.Param("id")), d, 2)
		c.JSON(200, gin.H{
			"status":  true,
			"message": "",
		})
		return
	}
	c.JSON(400, gin.H{
		"status":  false,
		"message": err,
	})
}
func DeviceLambdaDeployHandler(c *gin.Context) {
	var (
		evt = struct {
			Name    string   `json:"name"`
			Version string   `json:"version"`
			Cmd     []string `json:"cmd"`
			Args    []string `json:"args"`
			Env     []string `json:"env"`
		}{}
		err error
		buf []byte
	)
	buf, err = ioutil.ReadAll(c.Request.Body)
	if err == nil {
		json.Unmarshal(buf, &evt)
		p := pb.ThingMSG{
			MsgId:   111,
			Cid:     "test",
			MsgType: pb.ThingMSG_LAMBDA,
			Func:    fmt.Sprintf("%s:%s", evt.Name, evt.Version),
			Cmd:     []string{},
			Args:    evt.Args,
			Env:     evt.Env,
		}
		d, _ := proto.Marshal(&p)
		SendMSG(fmt.Sprintf("system/devices/%s", c.Param("id")), d, 2)
		c.JSON(200, gin.H{
			"status":  true,
			"message": "",
		})
		return
	}
	c.JSON(400, gin.H{
		"status":  false,
		"message": err,
	})
}

func SendMSG(topic string, d []byte, qos int) {
	t := mc.Publish(topic, byte(qos), false, d)
	t.Wait()
}

func DeviceMFT(sn string, name string) (string, error) {
	var device Device
	device.SN = sn
	if err := db.First(&device, device).Error; err != nil {
		u := uuid.NewHash(sha256.New(), uuid.NameSpaceDNS, []byte(sn), 5)
		device.ID = u
		device.Name = name
		db.Save(&device)
		return u.String(), nil
	}
	return device.ID.String(), nil
}
