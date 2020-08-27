package main

import (
	"../../proto"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net"
	"os"
	//"github.com/golang/protobuf/proto"
	"github.com/coreos/go-iptables"
	"github.com/vishvananda/netlink"
	"github.com/vishvananda/netns"
	"os/exec"
)

type RuncSpec struct {
	OciVersion string `json:"ociVersion"`

	InstanceID string `json:"instanceID"`
	Process    struct {
		Terminal bool `json:"terminal"`
		User     struct {
			UID int `json:"uid"`
			GID int `json:"gid"`
		} `json:"user"`
		Args         []string `josn:"args"`
		Env          []string `json:"env"`
		CWD          string   `json:"cwd"`
		Capabilities struct {
			Bounding   []string `json:"bounding"`
			Effective  []string `json:"effective"`
			Inhertable []string `json:"inhertable"`
			Permitted  []string `json:"permitted"`
			Ambient    []string `json:"ambient"`
		} `json:"capabilities"`
		RLimits []struct {
			Type string `json:"type"`
			Hard int    `json:"hard"`
			Soft int    `json:"soft"`
		} `json:"rlimits"`
		NoNewPrivileges bool `json:"noNewPrivileges"`
	} `json:"process"`
	Root struct {
		Path     string `json:"path"`
		Readonly bool   `json:"readonly"`
	} `json:"root"`
	Hostname string `json:"hostname"`
	Mounts   []struct {
		Destination string   `json:"destination"`
		Type        string   `json:"type"`
		Source      string   `json:"source"`
		Options     []string `json:"options,omitempty"`
	} `json:"mounts"`
	Linux struct {
		Resources struct {
			Devices []struct {
				Allow  bool   `json:"allow"`
				Access string `json:"access"`
			} `json:"devices"`
		} `json:"resources"`
		Namespaces []struct {
			Type string `json:"type"`
			Path string `json:"path,omitempty"`
		} `json:"namespaces"`
		MaskedPaths   []string `json:"maskedPaths"`
		ReadonlyPaths []string `json:"readonlyPaths"`
	} `json:"linux"`
}

func LambdaPy(p pb.ThingMSG) {
	var rs RuncSpec
	//if fd, err := os.Open(fmt.Sprintf("%s/lambda/%s/config.json", BaseDir, p.Cid)); err == nil {
	//	buf, _ := ioutil.ReadAll(fd)
	//	var rs RuncSpec
	//	json.Unmarshal(buf, &rs)
	//	rs.LambdaStart()
	//	return
	//}
	fd, _ := os.Open(fmt.Sprintf("%s/lambda/config.json", BaseDir))
	buf, _ := ioutil.ReadAll(fd)
	json.Unmarshal(buf, &rs)
	rs.InstanceID = p.Cid
	rs.LambdaPrep()
	rs.LambdaNetworkInit()
	rs.LambdaStart()
}

func LambdaEnvSetup() {
	la := netlink.NewLinkAttrs()
	la.Name = "aurora0"
	br := &netlink.Bridge{LinkAttrs: la}
	err := netlink.LinkAdd(br)
	if err != nil {
		fmt.Printf("could not add %s: %v\n", la.Name, err)
	}
	gw, err := netlink.LinkByName("aurora0")
	if err != nil {
		fmt.Println("get br", err)
	}
	var ip []byte
	var addr netlink.Addr
	ip, addr.IPNet, _ = net.ParseCIDR("101.101.0.1/16")
	addr.IP = ip
	err = netlink.AddrAdd(gw, &addr)
	netlink.LinkSetUp(gw)

	iptables

	cmd := "iptables -A FORWARD -i aurora0 -o aurora0 -j ACCEPT"
	exec.Command("bash", "-c", cmd).Run()
	cmd = "iptables -A FORWARD -i aurora0 ! -o aurora0 -j ACCEPT"
	exec.Command("bash", "-c", cmd).Run()

	cmd = "iptables -t nat -A POSTROUTING -s 101.101.0.0/16 ! -o aurora0 -j MASQUERADE"
	exec.Command("bash", "-c", cmd).Run()
}

func (rs *RuncSpec) LambdaPrep() error {
	err := os.MkdirAll(BaseDir+"/lambda/"+rs.InstanceID+"/rootfs", 0755)
	if err != nil {
		fmt.Println(err)
	}
	err = os.MkdirAll(BaseDir+"/lambda/"+rs.InstanceID+"/upper", 0755)
	if err != nil {
		fmt.Println(err)
	}
	err = os.MkdirAll(BaseDir+"/lambda/"+rs.InstanceID+"/worker", 0755)
	if err != nil {
		fmt.Println(err)
	}

	rs.Process.Terminal = false
	rs.Process.Args = []string{"/usr/bin/python3", "/lambda/lambda.py", "--mqtt", "msg.icetears.com", "--port", "443"}

	lower := fmt.Sprintf("%s/rootfs:%s/lambda", BaseDir, BaseDir)
	upper := fmt.Sprintf("%s/lambda/%s/upper", BaseDir, rs.InstanceID)
	worker := fmt.Sprintf("%s/lambda/%s/worker", BaseDir, rs.InstanceID)
	//merge := fmt.Sprintf("%s/lambda/%s/rootfs", BaseDir, rs.InstanceID)

	var overlay = struct {
		Destination string   `json:"destination"`
		Type        string   `json:"type"`
		Source      string   `json:"source"`
		Options     []string `json:"options,omitempty"`
	}{
		Destination: "/",
		Type:        "overlay",
		Source:      "overlay",
		Options: []string{
			"lowerdir=" + lower,
			"upperdir=" + upper,
			"workdir=" + worker,
		},
	}
	rs.Mounts = append(rs.Mounts[:1], rs.Mounts[0])
	rs.Mounts[0] = overlay

	rs.Process.Env = append(rs.Process.Env, "LC_ALL=C.UTF-8")
	rs.Process.Env = append(rs.Process.Env, "LANG=C.UTF-8")

	//rs.Root.Path = "/home/xiao/work/iot/aurora/rootfs"
	//cmd := "mount -t overlay overlay -o lowerdir=" + lower + ",upperdir=" + upper + ",workdir=" + worker + " " + merge
	//o, err := exec.Command("bash", "-c", cmd).Output()
	//if err != nil {
	//	fmt.Println(string(o), cmd, err)
	//}
	rs.Linux.Namespaces[1].Path = "/var/run/netns/" + rs.InstanceID

	d, _ := json.Marshal(rs)
	err = ioutil.WriteFile(fmt.Sprintf("%s/lambda/%s/config.json", BaseDir, rs.InstanceID), d, 0644)
	if err != nil {
		fmt.Println(err)
	}
	return err
}

func (rs *RuncSpec) LambdaNetworkInit() error {
	la := netlink.NewLinkAttrs()
	ve := &netlink.Veth{
		LinkAttrs: la,
		PeerName:  "veth1",
	}
	ve.Name = "veth0"
	if err := netlink.LinkAdd(ve); err != nil {
		fmt.Println("addveth", err)
	}
	br, err := netlink.LinkByName("aurora0")
	if err != nil {
		LambdaEnvSetup()
	}
	LambdaEnvSetup()
	netlink.LinkSetMaster(ve, br)

	var addr netlink.Addr
	var ip []byte
	ip, addr.IPNet, _ = net.ParseCIDR("101.101.0.10/16")
	addr.IP = ip

	netlink.LinkSetUp(ve)
	ver, _ := netlink.LinkByName(ve.PeerName)

	orins, _ := netns.Get()

	ns, _ := netns.NewNamed(rs.InstanceID)
	defer ns.Close()
	lo, err := netlink.LinkByName("lo")
	netlink.LinkSetUp(ver)
	netlink.LinkSetUp(lo)
	netns.Set(orins)

	err = netlink.LinkSetNsFd(ver, int(ns))
	if err != nil {
		fmt.Println("eee", err.Error())
	}
	netns.Set(ns)
	err = netlink.AddrAdd(ver, &addr)
	if err != nil {
		fmt.Println("eee", err.Error())
	}
	netlink.LinkSetUp(ver)

	cmd := fmt.Sprintf("ip r add default via 101.101.0.1")
	exec.Command("bash", "-c", cmd).Run()

	netns.Set(orins)

	return err
}

func (rs *RuncSpec) LambdaStart() error {
	cmd := fmt.Sprintf("cd %s/lambda/%s && runc run %s 2>&1", BaseDir, rs.InstanceID, rs.InstanceID)
	o, err := exec.Command("bash", "-c", cmd).Output()
	if err != nil {
		fmt.Println(string(o), cmd, err)
	}
	fmt.Println(cmd)
	return nil
}
