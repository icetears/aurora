package main

import (
	"encoding/json"
	"fmt"
	"github.com/c-bata/go-prompt"
	"github.com/google/uuid"
	"github.com/icetears/aurora/pkg/requests"
	"github.com/icetears/aurora/pkg/ssl"
	"github.com/olekukonko/tablewriter"
	"io/ioutil"
	"net/http"
	"os"
	"os/exec"
	"strings"
	"time"
)

var (
	version  string
	revision string
)

func Executor(s string) {
	cmds := strings.Fields(s)
	if len(cmds) == 0 {
		return
	} else if cmds[0] == "quit" || s == "exit" {
		fmt.Println("Bye!")
		os.Exit(0)
		return
	}

	switch cmds[0] {
	case "list":
		/* Table
		data := [][]string{
			[]string{"A", "The Good", "500"},
			[]string{"B", "The Very very Bad Man", "288"},
			[]string{"C", "The Ugly", "120"},
			[]string{"D", "The Gopher", "800"},
		}*/

		type Board struct {
			Name string    `json:"name"`
			ID   uuid.UUID `json:"id"`
		}

		var (
			err error
			b   []Board
		)

		endpoint := "https://www.icetears.com/v1/devices"
		resp, err := http.Get(endpoint)
		if err != nil {
			fmt.Println("po", err)
			return
		}

		body, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			fmt.Println(err)
			return
		}

		err = json.Unmarshal(body, &b)
		if err != nil {
			fmt.Println(err)
			return
		}

		table := tablewriter.NewWriter(os.Stdout)
		table.SetHeader([]string{"ID", "Name", "Owner", "Lab"})
		table.SetCaption(true, "  "+time.Now().Format(time.RFC850))

		for _, v := range b {
			//fmt.Println(v.Ports)
			table.Append([]string{fmt.Sprintf("%s", v.ID.String()), v.Name, "(nxa22939) Xiaodong", "BeiJing"})
		}
		table.Render() // Send output
	case "delete":
		r := requests.New()
		endpoint := fmt.Sprintf("https://www.icetears.com/v1/devices/%s", cmds[1])
		r.Delete(endpoint, nil)

	case "console":
		if len(cmds) == 2 {
			Console(cmds[1])
		}
		fmt.Println("usage: console device-id")
	case "enroll":
		endpoint := "https://www.icetears.com/v1/device/enroll"
		r := requests.New()

		p, _ := ssl.GenerateKey("P256")
		csr := ssl.GenerateCsr(p.PrivateKey, "efefef800a", "aurora edge certificate")
		csr.Template.DNSNames = []string{"int.msg.net"}
		if d, err := csr.GenCsrPem(); err != nil {
			fmt.Println(err.Error())
		} else {
			r.Post(endpoint, d)
		}

		body := r.Response.Body
		fmt.Println(string(body), string(p.EncodeToPem()))
	default:
		cmd := fmt.Sprintf("%s 2>&1", s)
		o, err := exec.Command("bash", "-c", cmd).Output()
		if err != nil {
			fmt.Println(err.Error())
		}
		fmt.Println(string(o))
	}
	return
}

func Completer(d prompt.Document) []prompt.Suggest {
	switch d.CurrentLine() {
	case "hehe":
		var s []prompt.Suggest
		s = append(s, prompt.Suggest{"con", "console device-id"})
		s = append(s, prompt.Suggest{"hehess", "sdfdsak"})
		return s
	}
	return nil
}

func main() {
	cc := fmt.Sprintf(`
	#####################################
	#               X CLI               #
	#####################################

Please use 'exit' or 'Ctrl-D' to exit
	`)
	fmt.Println(cc)
	defer fmt.Println("Bye!")
	p := prompt.New(
		Executor,
		Completer,
		prompt.OptionTitle("xlab client"),
		prompt.OptionPrefix("x > "),
		prompt.OptionInputTextColor(prompt.Blue),
	)
	p.Run()
}
