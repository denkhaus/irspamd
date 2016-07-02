package rspamd

import (
	"bufio"
	"bytes"
	"errors"
	"io"
	"net"
	"regexp"
	"strconv"
	"strings"

	"github.com/denkhaus/tcgl/applog"
	"gopkg.in/pipe.v2"
)

var (
	rspamInfoRe = regexp.MustCompile(`(.+)\/(.+)\s(\d+)\s(.+)`)
	rspamMainRe = regexp.MustCompile(`^Metric:\s([a-z]+);\s(False|True);\s(\d{1,2}\.\d{1,2})\s\/\s(\d{1,2}\.\d{1,2})\s\/\s(\d{1,2}\.\d{1,2})$`)
	//rspamDetailsRe = regexp.MustCompile(`^(-?[0-9\.]*)\s([a-zA-Z0-9_]*)(\W*)([\w:\s-]*)`)
)

type Config struct {
	Ip      string
	Port    int
	Timeout int
}

type rspamdData struct {
	config   *Config
	RawEmail []byte
}

//type RspamdHeader struct {
//	Pts         string
//	RuleName    string
//	Description string
//}

type Response struct {
	ResponseCode    int
	ResponseMessage string
	Score           float64
	Spam            bool
	Threshold       float64
	//Details         []RspamdHeader
}

////////////////////////////////////////////////////////////////////////////////
func CheckSpam(config *Config, email []byte) (*Response, error) {
	rspamd := &rspamdData{
		config:   config,
		RawEmail: email,
	}

	output, err := rspamd.checkEmail()
	if err != nil {
		return nil, err
	}

	resp := rspamd.parseOutput(output)
	return resp, nil
}

////////////////////////////////////////////////////////////////////////////////
func (ss *rspamdData) checkEmail() ([]string, error) {

	ip := net.ParseIP(ss.config.Ip)
	if ip == nil {
		return nil, errors.New("Invalid ip address")
	}
	addr := &net.TCPAddr{
		IP:   ip,
		Port: ss.config.Port,
	}
	conn, err := net.DialTCP("tcp", nil, addr)
	if err != nil {
		return nil, err
	}
	defer conn.Close()
	// write headers
	_, err = conn.Write([]byte("CHECK RSPAMC/1.3\r\n"))
	if err != nil {
		return nil, err
	}
	_, err = conn.Write([]byte("Content-length: " + strconv.Itoa(len(ss.RawEmail)) + "\r\n\r\n"))
	if err != nil {
		return nil, err
	}
	// write email
	_, err = conn.Write(ss.RawEmail)
	if err != nil {
		return nil, err
	}
	// force close writer
	conn.CloseWrite()

	// read data
	var dataArrays []string
	reader := bufio.NewReader(conn)
	// reading
	for {
		line, err := reader.ReadString('\n')
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, err
		}

		line = strings.TrimRight(line, " \t\r\n")
		dataArrays = append(dataArrays, line)
	}

	return dataArrays, nil
}

// parse spamassassin output
////////////////////////////////////////////////////////////////////////////////
func (ss *rspamdData) parseOutput(output []string) *Response {
	response := &Response{}
	for _, row := range output {
		// header
		if rspamInfoRe.MatchString(row) {
			res := rspamInfoRe.FindStringSubmatch(row)
			if len(res) == 5 {
				if resCode, err := strconv.Atoi(res[3]); err == nil {
					response.ResponseCode = resCode
				}
				response.ResponseMessage = res[4]
			}
		}
		// summary
		if rspamMainRe.MatchString(row) {
			res := rspamMainRe.FindStringSubmatch(row)
			if len(res) == 6 {
				if strings.ToLower(res[2]) == "true" {
					response.Spam = true
				} else if strings.ToLower(res[2]) == "false" {
					response.Spam = false
				}
				if resFloat, err := strconv.ParseFloat(res[3], 64); err == nil {
					response.Score = resFloat
				}
				if resFloat, err := strconv.ParseFloat(res[4], 64); err == nil {
					response.Threshold = resFloat
				}
			}
		}
		// details
		//row = strings.Trim(row, " \t\r\n")
		//if rspamDetailsRe.MatchString(row) {
		//	res := spamDetailsRe.FindStringSubmatch(row)
		//	if len(res) == 5 {
		//		header := RspamdHeader{Pts: res[1], RuleName: res[2], Description: res[4]}
		//		response.Details = append(response.Details, header)
		//	}
		//}
	}
	return response
}

////////////////////////////////////////////////////////////////////////////////
func Learn(fnString string, reader io.Reader) (string, error) {
	b := &bytes.Buffer{}
	p := pipe.Line(
		pipe.Read(reader),
		pipe.Exec("rspamc", fnString),
		pipe.Write(b),
	)

	if err := pipe.Run(p); err != nil {
		return "", err
	}

	res := b.String()
	applog.Infof("Rspamd::%s::result::%s", fnString, res)
	return res, nil
}

////////////////////////////////////////////////////////////////////////////////
func LearnSpam(reader io.Reader) (string, error) {
	return Learn("learn_spam", reader)
}

////////////////////////////////////////////////////////////////////////////////
func LearnHam(reader io.Reader) (string, error) {
	return Learn("learn_ham", reader)
}
