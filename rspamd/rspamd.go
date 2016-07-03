package rspamd

import (
	"bytes"
	"fmt"
	"io"
	"regexp"
	"strconv"
	"strings"

	"github.com/denkhaus/tcgl/applog"
	"github.com/juju/errors"
	"gopkg.in/pipe.v2"
)

var (
	rspamSpamRe      = regexp.MustCompile(`^Spam:\s(false|true)$`)
	rspamScoreRe     = regexp.MustCompile(`^Score:\s(\d{1,2}\.\d{1,2})\s\/\s(\d{1,2}\.\d{1,2})$`)
	rspamMessageIdRe = regexp.MustCompile(`^Message-ID:\s(.*)$`)
	rspamTookTimeRe  = regexp.MustCompile(`^Results for file: stdin \((\d{1,2}\.\d{1,3})\sseconds\)$`)
	rspamLearnRespRe = regexp.MustCompile(`^HTTP error:\s(\d{3}),\s<(.*)>\s(.*)$`)
)

type CheckResponse struct {
	Message   string
	Score     float64
	Spam      bool
	Threshold float64
	Took      float64
	MessageId string
}

////////////////////////////////////////////////////////////////////////////////
func (p CheckResponse) FmtScore(pattern string) string {
	x := int(p.Score/p.Threshold*100) / 10 * 10
	return fmt.Sprintf(`%s%d`, pattern, x)
}

////////////////////////////////////////////////////////////////////////////////
func (p CheckResponse) Report(uid uint32) {
	if p.Spam {
		applog.Infof("mail %d is marked as spam. score %f, took %f sec", uid, p.Score, p.Took)
	} else {
		applog.Infof("mail %d is clean, took %f sec", uid, p.Took)
	}
}

type LearnResponse struct {
	Message        string
	Success        bool
	Skiped         bool
	Took           float64
	ErrorCode      int
	ErrorMessageID string
	ErrorMessage   string
}

////////////////////////////////////////////////////////////////////////////////
func (p LearnResponse) Report(uid uint32) {
	if p.Success {
		applog.Infof("mail %d learned successfully in %f sec", uid, p.Took)
		return
	}

	if p.Skiped {
		applog.Infof("mail %d: already learned, took %f sec", uid, p.Took)
		return
	}

	applog.Infof("mail %d: not learned: %s", uid, p.ErrorMessage)
}

////////////////////////////////////////////////////////////////////////////////
func parseLines(reader *bytes.Buffer) ([]string, error) {
	var data []string

	for {
		line, err := reader.ReadString('\n')
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, err
		}

		line = strings.TrimRight(line, " \t\r\n")
		data = append(data, line)
	}
	return data, nil
}

////////////////////////////////////////////////////////////////////////////////
func parseLearnOutput(reader *bytes.Buffer) (*LearnResponse, error) {
	data, err := parseLines(reader)
	if err != nil {
		return nil, errors.Annotate(err, "parse lines")
	}

	response := &LearnResponse{
		Message: data[0],
	}

	for _, row := range data {
		if row == "success = true;" {
			response.Success = true
		}
		if rspamTookTimeRe.MatchString(row) {
			res := rspamTookTimeRe.FindStringSubmatch(row)
			if resFloat, err := strconv.ParseFloat(res[1], 64); err == nil {
				response.Took = resFloat
			}
		}
		if rspamLearnRespRe.MatchString(row) {
			res := rspamLearnRespRe.FindStringSubmatch(row)
			if resInt, err := strconv.Atoi(res[1]); err == nil {
				response.ErrorCode = resInt
			}

			response.ErrorMessageID = res[2]
			response.ErrorMessage = res[3]
		}
	}

	if response.ErrorCode == 404 &&
		response.ErrorMessage == "has been already learned as spam, ignore it" {
		response.Skiped = true
	}
	return response, nil

}

////////////////////////////////////////////////////////////////////////////////
func parseCheckOutput(reader *bytes.Buffer) (*CheckResponse, error) {
	data, err := parseLines(reader)
	if err != nil {
		return nil, errors.Annotate(err, "parse lines")
	}

	response := &CheckResponse{
		Message: data[0],
	}

	for _, row := range data {
		if rspamScoreRe.MatchString(row) {
			res := rspamScoreRe.FindStringSubmatch(row)
			if resFloat, err := strconv.ParseFloat(res[1], 64); err == nil {
				response.Score = resFloat
			}
			if resFloat, err := strconv.ParseFloat(res[2], 64); err == nil {
				response.Threshold = resFloat
			}
		}
		if rspamSpamRe.MatchString(row) {
			res := rspamSpamRe.FindStringSubmatch(row)

			if strings.ToLower(res[1]) == "true" {
				response.Spam = true
			} else if strings.ToLower(res[1]) == "false" {
				response.Spam = false
			}
		}
		if rspamMessageIdRe.MatchString(row) {
			res := rspamMessageIdRe.FindStringSubmatch(row)
			response.MessageId = res[1]
		}
		if rspamTookTimeRe.MatchString(row) {
			res := rspamTookTimeRe.FindStringSubmatch(row)
			if resFloat, err := strconv.ParseFloat(res[1], 64); err == nil {
				response.Took = resFloat
			}
		}
	}
	return response, nil
}

////////////////////////////////////////////////////////////////////////////////
func Check(reader io.Reader) (*CheckResponse, error) {
	b := &bytes.Buffer{}
	p := pipe.Line(
		pipe.Read(reader),
		pipe.Exec("rspamc"),
		pipe.Write(b),
	)

	if err := pipe.Run(p); err != nil {
		return nil, err
	}

	return parseCheckOutput(b)
}

////////////////////////////////////////////////////////////////////////////////
func Learn(fnString string, reader io.Reader) (*LearnResponse, error) {
	b := &bytes.Buffer{}
	p := pipe.Line(
		pipe.Read(reader),
		pipe.Exec("rspamc", fnString),
		pipe.Write(b),
	)

	if err := pipe.Run(p); err != nil {
		return nil, err
	}

	return parseLearnOutput(b)
}

////////////////////////////////////////////////////////////////////////////////
func LearnSpam(reader io.Reader) (*LearnResponse, error) {
	return Learn("learn_spam", reader)
}

////////////////////////////////////////////////////////////////////////////////
func LearnHam(reader io.Reader) (*LearnResponse, error) {
	return Learn("learn_ham", reader)
}
