package schtab

import (
	"bufio"
	"bytes"
	"encoding/csv"
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
	"strconv"
	"strings"
)

const SCHTASKS_EXE = "schtasks.exe"

const (
	// Sub command flags
	FL_QUERY  = "/query"
	FL_CREATE = "/create"
	FL_DELETE = "/delete"
	FL_CHANGE = "/change"
	FL_RUN    = "/run"
	FL_END    = "/end"

	// Global flags
	FL_TASKNAME = "/tn"
	FL_FORCE    = "/f"

	// Create flags
	FL_SCHEDULE  = "/sc"
	FL_TASKRUN   = "/tr"
	FL_MODIFIERS = "/mo"
	FL_DAY       = "/d"
	FL_MONTH     = "/m"
	FL_STARTTIME = "/st"

	// Query flags
	FL_FORMAT   = "/fo"
	FL_NOHEADER = "/nh"
	FL_VERBOSE  = "/v"
)

const (
	// Schedule types
	SC_MINUTE  = "MINUTE"
	SC_HOURLY  = "HOURLY"
	SC_DAILY   = "DAILY"
	SC_WEEKLY  = "WEEKLY"
	SC_MONTHLY = "MONTHLY"
	SC_ONCE    = "ONCE"
	SC_ONSTART = "ONSTART"
	SC_ONLOGON = "ONLOGON"
	SC_ONIDLE  = "ONIDLE"
)

const (
	// Query formats
	FO_TABLE = "TABLE"
	FO_LIST  = "LIST"
	FO_CSV   = "CSV"
)

const (
	// Task name parts
	TN_ROOT = "schtab"
	TN_BASE = "task-%03d"
)

func getTnPrefix() string {
	return `\` + TN_ROOT + `\` + os.Getenv("USERNAME") + `\`
}

func atomon(s string) string {
	return map[string]string{
		"1":  "JAN",
		"2":  "FEB",
		"3":  "MAR",
		"4":  "APR",
		"5":  "MAY",
		"6":  "JUN",
		"7":  "JUL",
		"8":  "AUG",
		"9":  "SEP",
		"10": "OCT",
		"11": "NOV",
		"12": "DEC",
	}[s]
}

func atodow(s string) string {
	return map[string]string{
		"0": "SUN",
		"1": "MON",
		"2": "TUE",
		"3": "WED",
		"4": "THU",
		"5": "FRI",
		"6": "SAT",
	}[s]
}

// RegisterAll registers all tasks in crontab format text at Task Scheduler
func RegisterAll(r io.Reader) error {
	sc := bufio.NewScanner(r)
	i := 1
	for sc.Scan() {
		ts := strings.TrimSpace(sc.Text())
		if ts == "" || strings.HasPrefix(ts, "#") {
			continue
		}
		t := NewTask(getTnPrefix() + fmt.Sprintf(TN_BASE, i))
		if err := t.SetCron(ts); err != nil {
			fmt.Fprintln(os.Stderr, "Error:", err)
			continue
		}
		if err := t.Register(); err != nil {
			return err
		}
		i++
	}
	return nil
}

// UnregisterAll unregister all current user's tasks registered by schtab from Task Scheduler
func UnregisterAll() error {
	rt, err := RegisteredTasks(getTnPrefix())
	if err != nil {
		return err
	}
	for _, v := range rt {
		if err := v.Unregister(); err != nil {
			return err
		}
	}
	return nil
}

type Task struct {
	tn string
	tr string
	sc string
	mo string
	d  string
	m  string
	st string
}

// NewTask returns a new task
func NewTask(name string) *Task {
	return &Task{
		tn: strings.ToLower(name),
	}
}

// RegisteredTasks returns all tasks with the specified prefix registered at Task Scheduler
func RegisteredTasks(prfx string) ([]Task, error) {
	out, err := exec.Command(SCHTASKS_EXE, FL_QUERY, FL_FORMAT, FO_CSV, FL_NOHEADER).Output()
	if err != nil {
		return nil, err
	}
	rec, err := csv.NewReader(bytes.NewBuffer(out)).ReadAll()
	if err != nil {
		return nil, err
	}
	p := strings.ToLower(prfx)
	t := []Task{}
	for _, r := range rec {
		if strings.HasPrefix(r[0], p) {
			t = append(t, *NewTask(r[0]))
		}
	}
	return t, nil
}

// Register registers this task at Task Scheduler
func (t *Task) Register() error {
	if t.tn == "" || t.tr == "" || t.sc == "" {
		return errors.New("schedule is unset yet")
	}
	args := []string{FL_CREATE, FL_FORCE, FL_TASKNAME, t.tn, FL_TASKRUN, t.tr, FL_SCHEDULE, t.sc}
	if t.mo != "" {
		args = append(args, FL_MODIFIERS, t.mo)
	}
	if t.d != "" {
		args = append(args, FL_DAY, t.d)
	}
	if t.m != "" {
		args = append(args, FL_MONTH, t.m)
	}
	if t.st != "" {
		args = append(args, FL_STARTTIME, t.st)
	}
	if err := exec.Command(SCHTASKS_EXE, args...).Run(); err != nil {
		return err
	}
	return nil
}

// Unregister unregisters this task from Task Scheduler
func (t *Task) Unregister() error {
	if t.tn == "" {
		return errors.New("task name is unset yet")
	}
	if err := exec.Command(SCHTASKS_EXE, FL_DELETE, FL_FORCE, FL_TASKNAME, t.tn).Run(); err != nil {
		return err
	}
	return nil
}

// SetCron sets schedule setting from a row of crontab format text
func (t *Task) SetCron(s string) error {
	f := strings.Fields(s)
	if len(f) < 6 {
		return errors.New("too few fields")
	}

	// Set command
	t.tr = strings.Join(f[5:], " ")

	// Set schedule type
	if f[0] == "*" {
		t.sc = SC_MINUTE
	} else if f[1] == "*" {
		t.sc = SC_HOURLY
	} else if f[4] != "*" {
		t.sc = SC_WEEKLY
	} else if f[2] == "*" {
		t.sc = SC_DAILY
	} else if f[3] == "*" {
		t.sc = SC_MONTHLY
	} else {
		t.sc = SC_MONTHLY
		t.m = atomon(f[3])
	}

	// Set day
	if f[4] != "*" {
		t.d = atodow(f[4])
	} else if f[2] != "*" {
		t.d = f[2]
	}

	// Set start time
	m, err := strconv.Atoi(f[0])
	if err != nil {
		m = 0
	}
	h, err := strconv.Atoi(f[1])
	if err != nil {
		h = 0
	}
	t.st = fmt.Sprintf("%02d:%02d", h, m)

	return nil
}
