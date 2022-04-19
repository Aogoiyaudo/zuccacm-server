package config

import (
	"bytes"
	"fmt"
	"os"
	"path"
	"sort"
	"strings"

	log "github.com/sirupsen/logrus"

	"zuccacm-server/utils"
)

type MyFormatter struct {
	colorful bool
}

const (
	red    = 31
	yellow = 33
	blue   = 36
	gray   = 37
)

func (f *MyFormatter) color(s string, c int) string {
	if f.colorful {
		return fmt.Sprintf("\x1b[%dm%s\x1b[0m", c, s)
	} else {
		return s
	}
}

func (f *MyFormatter) Format(entry *log.Entry) ([]byte, error) {
	var levelColor int
	switch entry.Level {
	case log.DebugLevel, log.TraceLevel:
		levelColor = gray
	case log.WarnLevel:
		levelColor = yellow
	case log.ErrorLevel, log.PanicLevel, log.FatalLevel:
		levelColor = red
	case log.InfoLevel:
		levelColor = blue
	default:
		levelColor = blue
	}

	var b *bytes.Buffer
	if entry.Buffer != nil {
		b = entry.Buffer
	} else {
		b = &bytes.Buffer{}
	}
	level := strings.ToUpper(entry.Level.String()[:4])
	datetime := entry.Time.Format("2006-01-02 15:04:05")
	file := fmt.Sprintf("%s:%d", utils.SimplePath(entry.Caller.File, RootDir), entry.Caller.Line)
	function := fmt.Sprintf("[%s]", path.Base(entry.Caller.Function))
	msg := entry.Message
	out := fmt.Sprintf("%s[%s]  %s", f.color(level, levelColor), datetime, file)
	out = fmt.Sprintf("%-55s  %s", out, function)
	out = fmt.Sprintf("%-85s  %s", out, msg)
	out = fmt.Sprintf("%-135s  ", out)
	type Data struct {
		k string
		v interface{}
	}
	var data []Data
	for k, v := range entry.Data {
		data = append(data, Data{k, v})
	}
	sort.Slice(data, func(i, j int) bool {
		return data[i].k < data[j].k
	})
	for _, x := range data {
		out = fmt.Sprintf("%s %s=%v", out, f.color(x.k, levelColor), x.v)
	}
	b.WriteString(out + "\n")
	return b.Bytes(), nil
}

func SetLogForm(colorful bool) {
	log.SetReportCaller(true)
	log.SetFormatter(&MyFormatter{colorful: colorful})
}

func initLog() (path string, level string) {
	if file, err := os.OpenFile(Instance.Path, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0666); err != nil {
		path = "os.Stdout"
		SetLogForm(true)
		log.SetOutput(os.Stdout)
		log.WithFields(log.Fields{
			"path":  Instance.Path,
			"error": err,
		}).Warnf("Parse log path failed! Use os.Stdout instead.")
	} else {
		path = Instance.Path
		SetLogForm(false)
		log.SetOutput(file)
	}
	if level, err := log.ParseLevel(Instance.Level); err != nil {
		log.SetLevel(log.InfoLevel)
		log.WithFields(log.Fields{
			"level": Instance.Level,
			"error": err,
		}).Warnf("Parse log level failed! Use info Level instead.")
	} else {
		log.SetLevel(level)
	}
	level = log.GetLevel().String()
	return
}
