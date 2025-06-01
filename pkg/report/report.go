package report

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/thd3r/SimpHttp/pkg/utils"
)

type DataOutput struct {
	Proto    string  `json:"proto"`
	Host     string  `json:"host"`
	Status   *string `json:"status"`
	SizeBody *string `json:"size_body"`
	Redirect *string `json:"redirected_to"`
	ErrorMsg *string `json:"error_message"`
}

type DataReport struct {
	Info      string       `json:"info"`
	Version   string       `json:"version"`
	Timestamp time.Time    `json:"timestamp"`
	Output    []DataOutput `json:"data_output"`
}

var FilePath = fmt.Sprintf("%s/SimpHttp-%v.json", os.TempDir(), time.Now().UnixNano())

func JsonReport(verbose bool, data <-chan DataOutput) {
	file, err := os.OpenFile(FilePath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
	if err != nil {
		utils.VerbosePrint(verbose, fmt.Sprintf("%s\t%s\n", utils.ColoredText("red", "ERR"), strings.ToLower(err.Error())))
	}
	defer file.Close()

	var results DataReport

	results.Info = "SimpHttp-Output"
	results.Version = utils.CurrentVersion
	results.Timestamp = time.Now()

	for d := range data {
		results.Output = append(results.Output, DataOutput{
			Proto:    d.Proto,
			Host:     d.Host,
			Status:   d.Status,
			SizeBody: d.SizeBody,
			Redirect: d.Redirect,
			ErrorMsg: d.ErrorMsg,
		})
	}

	encoder := json.NewEncoder(file)
	encoder.SetIndent("", "  ")
	encoder.SetEscapeHTML(false)

	if err := encoder.Encode(results); err != nil {
		utils.VerbosePrint(verbose, fmt.Sprintf("%s\t%s\n", utils.ColoredText("red", "error"), strings.ToLower(err.Error())))
	}

	if len(results.Output) > 0 {
		fmt.Printf("\n:: Report saved to %s\n", FilePath)
	}
}
