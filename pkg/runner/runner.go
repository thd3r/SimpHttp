package runner

import (
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/thd3r/SimpHttp/pkg/net"
	"github.com/thd3r/SimpHttp/pkg/net/client"
	"github.com/thd3r/SimpHttp/pkg/report"
	"github.com/thd3r/SimpHttp/pkg/utils"
)

type SimpHttpBase struct {
	Targets []string
	Threads int
	Timeout int
	Verbose bool
	Client  *client.Client
}

func NewSimpHttp(target string, threads, timeout int, verbose bool) *SimpHttpBase {
	var targets []string

	if target != "" {
		if utils.IsFile(target) {
			file, err := os.Open(target)
			if err != nil {
				fmt.Printf("%s: failed to open file: %v\n", utils.ColoredText("red", "eror"), err)
				os.Exit(1)
			}
			defer file.Close()

			lines, err := utils.ReadLines(file)
			if err != nil {
				fmt.Printf("%s: failed to read lines from file: %v\n", utils.ColoredText("red", "eror"), err)
				os.Exit(1)
			}

			targets = append(targets, lines...)
		} else {
			targets = append(targets, target)
		}
	} else {
		lines, err := utils.ReadLines(os.Stdin)
		if err != nil {
			fmt.Printf("%s: failed to read from stdin: %v\n", utils.ColoredText("red", "eror"), err)
		}
		targets = append(targets, lines...)
	}

	if len(targets) == 0 {
		fmt.Printf("%s: no targets provided\n", utils.ColoredText("red", "eror"))
		os.Exit(1)
	}

	clients := client.NewClient(timeout)

	return &SimpHttpBase{
		Targets: targets,
		Threads: threads,
		Timeout: timeout,
		Verbose: verbose,
		Client:  clients,
	}
}

func (base *SimpHttpBase) SimpHttpRun() {
	urlsChan := make(chan string)
	hostsChan := make(chan string)
	validHostsChan := make(chan string)
	dataOutputChan := make(chan report.DataOutput)

	var validateHostWG sync.WaitGroup
	for i := 0; i < base.Threads; i++ {
		validateHostWG.Add(1)

		go func() {
			defer validateHostWG.Done()

			for host := range hostsChan {
				base.validateHost(host, validHostsChan)
			}
		}()
	}

	var httpRunWG sync.WaitGroup
	for i := 0; i < base.Threads; i++ {
		httpRunWG.Add(1)

		go func() {
			defer httpRunWG.Done()

			for validHostsChan != nil || urlsChan != nil {
				select {
				case host, ok := <-validHostsChan:
					if !ok {
						validHostsChan = nil
						continue
					}
					base.httpRobeWorker(host, dataOutputChan)
				case url, ok := <-urlsChan:
					if !ok {
						urlsChan = nil
						continue
					}
					base.httpBasicWorker(url, dataOutputChan)
				}
			}

		}()
	}

	go func() {
		for _, target := range base.Targets {
			if base.isUrl(target) {
				urlsChan <- target
			} else {
				hostsChan <- target
			}
		}
		close(urlsChan)
		close(hostsChan)
	}()

	go func() {
		validateHostWG.Wait()
		close(validHostsChan)
	}()

	go func() {
		httpRunWG.Wait()
		close(dataOutputChan)
	}()

	report.JsonReport(base.Verbose, dataOutputChan)
}

func (base *SimpHttpBase) httpRobeWorker(host string, dataOutputChan chan<- report.DataOutput) {
	for _, proto := range []string{"http", "https"} {
		utils.VerbosePrint(base.Verbose, fmt.Sprintf("%s: processing %s for %s\n", utils.ColoredText("blue", "info"), strings.ToUpper(proto), host))

		resp, err := base.Client.Do("GET", fmt.Sprintf("%s://%s", proto, host))
		if err != nil {
			errMsg := fmt.Sprintf("fetching %s — %v", host, err)
			utils.VerbosePrint(base.Verbose, fmt.Sprintf("%s: %s\n", utils.ColoredText("red", "eror"), errMsg))
			dataOutputChan <- report.DataOutput{
				Url:      fmt.Sprintf("%s://%s", proto, host),
				Proto:    strings.ToUpper(proto),
				Host:     host,
				ErrorMsg: errMsg,
				IsError:  true,
			}
			if proto == "http" {
				continue // try HTTPS if HTTP fails
			}
			break // stop if HTTPS also fails
		}

		size, err := io.Copy(io.Discard, resp.Body)
		if err != nil {
			errMsg := fmt.Sprintf("reading response body for %s — %v", host, err)
			utils.VerbosePrint(base.Verbose, fmt.Sprintf("%s: %s\n", utils.ColoredText("red", "eror"), errMsg))
			dataOutputChan <- report.DataOutput{
				Url:      fmt.Sprintf("%s://%s", proto, host),
				Proto:    strings.ToUpper(proto),
				Host:     host,
				Status:   resp.Status,
				ErrorMsg: errMsg,
				IsError:  true,
			}
			if proto == "http" {
				continue // try HTTPS if reading HTTP body fails
			}
			break // stops if reading the HTTPS body fails
		}

		sizeBody := fmt.Sprintf("%dw", size)
		base.processResponse(resp, host, proto, sizeBody, dataOutputChan)

		resp.Body.Close()
		break // stop after success
	}
}

func (base *SimpHttpBase) httpBasicWorker(url string, dataOutputChan chan<- report.DataOutput) {
	utils.VerbosePrint(base.Verbose, fmt.Sprintf("%s: processing %s\n", utils.ColoredText("blue", "info"), url))

	parse, err := base.parseUrl(url)
	if err != nil {
		utils.VerbosePrint(base.Verbose, fmt.Sprintf("%s: %v\n", utils.ColoredText("red", "eror"), err))
		return
	}

	resp, err := base.Client.Do("GET", url)
	if err != nil {
		errMsg := fmt.Sprintf("fetching %s — %v", url, err)
		utils.VerbosePrint(base.Verbose, fmt.Sprintf("%s: %s\n", utils.ColoredText("red", "eror"), errMsg))
		dataOutputChan <- report.DataOutput{
			Url:      url,
			Proto:    strings.ToUpper(parse.Scheme),
			Host:     parse.Host,
			ErrorMsg: errMsg,
			IsError:  true,
		}
		return
	}
	defer resp.Body.Close()

	size, err := io.Copy(io.Discard, resp.Body)
	if err != nil {
		errMsg := fmt.Sprintf("reading response body for %s — %v", url, err)
		utils.VerbosePrint(base.Verbose, fmt.Sprintf("%s: %s\n", utils.ColoredText("red", "eror"), errMsg))
		dataOutputChan <- report.DataOutput{
			Url:      url,
			Proto:    strings.ToUpper(parse.Scheme),
			Host:     parse.Host,
			ErrorMsg: errMsg,
			IsError:  true,
		}
		return
	}

	sizeBody := fmt.Sprintf("%dw", size)
	base.processResponse(resp, parse.Host, parse.Scheme, sizeBody, dataOutputChan)
}

func (base *SimpHttpBase) validateHost(host string, validHostsChan chan<- string) {
	var once sync.Once

	for _, port := range []string{"80", "443"} {
		if net.IsReachableHost(host, port, time.Duration(3*time.Second)) {
			utils.VerbosePrint(base.Verbose, fmt.Sprintf("%s: valid host %s with port %s\n", utils.ColoredText("blue", "info"), host, port))

			once.Do(func() {
				validHostsChan <- host
			})

		} else {
			utils.VerbosePrint(base.Verbose, fmt.Sprintf("%s: invalid host %s with port %s\n", utils.ColoredText("red", "eror"), host, port))
		}
	}
}

func (base *SimpHttpBase) processResponse(resp *http.Response, host, proto, sizeBody string, dataOutputChan chan<- report.DataOutput) {
	switch {
	case resp.StatusCode >= 200 && resp.StatusCode < 300:
		fmt.Printf("%s %s %s\n", resp.Request.URL, utils.ColoredText("green", strconv.Itoa(resp.StatusCode)), utils.ColoredText("gray", sizeBody))
		dataOutputChan <- report.DataOutput{
			Url:      fmt.Sprintf("%s://%s", proto, host),
			Proto:    strings.ToUpper(proto),
			Host:     host,
			Status:   resp.Status,
			SizeBody: sizeBody,
		}
	case resp.StatusCode >= 300 && resp.StatusCode < 400:
		location := resp.Header.Get("Location")

		if !strings.HasPrefix(location, "http") {
			fmt.Printf("%s %s %s => %s\n", resp.Request.URL, utils.ColoredText("blue", strconv.Itoa(resp.StatusCode)), utils.ColoredText("gray", sizeBody), utils.ColoredText("cyan", fmt.Sprintf("%s%s", resp.Request.URL, location)))
			dataOutputChan <- report.DataOutput{
				Url:      fmt.Sprintf("%s://%s", proto, host),
				Proto:    strings.ToUpper(proto),
				Host:     host,
				Status:   resp.Status,
				SizeBody: sizeBody,
				Redirect: location,
			}
		} else {
			fmt.Printf("%s %s %s => %s\n", resp.Request.URL, utils.ColoredText("blue", strconv.Itoa(resp.StatusCode)), utils.ColoredText("gray", sizeBody), utils.ColoredText("cyan", location))
			dataOutputChan <- report.DataOutput{
				Url:      fmt.Sprintf("%s://%s", proto, host),
				Proto:    strings.ToUpper(proto),
				Host:     host,
				Status:   resp.Status,
				SizeBody: sizeBody,
				Redirect: location,
			}
		}
	case resp.StatusCode >= 400 && resp.StatusCode < 500:
		fmt.Printf("%s %s %s\n", resp.Request.URL, utils.ColoredText("magenta", strconv.Itoa(resp.StatusCode)), utils.ColoredText("gray", sizeBody))
		dataOutputChan <- report.DataOutput{
			Url:      fmt.Sprintf("%s://%s", proto, host),
			Proto:    strings.ToUpper(proto),
			Host:     host,
			Status:   resp.Status,
			SizeBody: sizeBody,
		}
	case resp.StatusCode >= 500 && resp.StatusCode < 600:
		fmt.Printf("%s %s %s\n", resp.Request.URL, utils.ColoredText("yellow", strconv.Itoa(resp.StatusCode)), utils.ColoredText("gray", sizeBody))
		dataOutputChan <- report.DataOutput{
			Url:      fmt.Sprintf("%s://%s", proto, host),
			Proto:    strings.ToUpper(proto),
			Host:     host,
			Status:   resp.Status,
			SizeBody: sizeBody,
		}
	}
}

func (base *SimpHttpBase) parseUrl(u string) (*url.URL, error) {
	parse, err := url.Parse(u)
	if err != nil {
		return nil, err
	}

	return parse, nil
}

func (base *SimpHttpBase) isUrl(u string) bool {
	parse, err := base.parseUrl(u)
	if err != nil {
		return false
	}

	return parse.Scheme == "http" || parse.Scheme == "https"
}
