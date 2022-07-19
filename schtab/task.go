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

// schtasks.exe
const SCHTASKS_EXE = "schtasks.exe"

// Column index No. for Task Name in Query command
const QUERY_TN_IDX = 0

const (
	// Sub command flags
	FLG_QUERY  = "/query"
	FLG_CREATE = "/create"
	FLG_DELETE = "/delete"
	FLG_CHANGE = "/change"
	FLG_RUN    = "/run"
	FLG_END    = "/end"

	// Global flags
	FLG_TASKNAME = "/tn"
	FLG_FORCE    = "/f"

	// Create flags
	FLG_SCHEDULE  = "/sc"
	FLG_TASKRUN   = "/tr"
	FLG_MODIFIERS = "/mo"
	FLG_DAY       = "/d"
	FLG_MONTH     = "/m"
	FLG_STARTTIME = "/st"

	// Query flags
	FLG_FORMAT   = "/fo"
	FLG_NOHEADER = "/nh"
	FLG_VERBOSE  = "/v"
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
	// Task name
	TN_ROOT = "schtab"
	TN_BASE = "task-%03d"
)

const (
	// crontab fields
	FLD_MINUTE = iota
	FLD_HOUR
	FLD_DAY
	FLD_MONTH
	FLD_WEEKDAY
	FLD_COMMAND
)

func dtstoi(s string) (int, error) {
	i, err := strconv.Atoi(s)
	if err == nil {
		return i, nil
	}
	i, ok := map[string]int{
		"JAN": 1,
		"FEB": 2,
		"MAR": 3,
		"APR": 4,
		"MAY": 5,
		"JUN": 6,
		"JUL": 7,
		"AUG": 8,
		"SEP": 9,
		"OCT": 10,
		"NOV": 11,
		"DEC": 12,
		"SUN": 0,
		"MON": 1,
		"TUE": 2,
		"WED": 3,
		"THU": 4,
		"FRI": 5,
		"SAT": 6,
	}[strings.ToUpper(s)]
	if !ok {
		return -1, errors.New("invalid literal")
	}
	return i, nil
}

func itomon(i int) string {
	return map[int]string{
		1:  "JAN",
		2:  "FEB",
		3:  "MAR",
		4:  "APR",
		5:  "MAY",
		6:  "JUN",
		7:  "JUL",
		8:  "AUG",
		9:  "SEP",
		10: "OCT",
		11: "NOV",
		12: "DEC",
	}[i]
}

func itodow(i int) string {
	return map[int]string{
		0: "SUN",
		1: "MON",
		2: "TUE",
		3: "WED",
		4: "THU",
		5: "FRI",
		6: "SAT",
		7: "SUN",
	}[i]
}

func getTnPrefix() string {
	return `\` + TN_ROOT + `\` + os.Getenv("USERNAME") + `\`
}

// RegisterAll registers all tasks in crontab format text at Task Scheduler
func RegisterAll(r io.Reader) error {
	sc := bufio.NewScanner(r)
	no := 1
	repl := ""
	for i := 1; sc.Scan(); i++ {
		ts := strings.TrimSpace(sc.Text())
		// Skip blank row
		if ts == "" {
			continue
		}
		// Replace task name
		if strings.HasPrefix(ts, "#") {
			uncom := strings.TrimSpace(strings.TrimPrefix(ts, "#"))
			if strings.HasPrefix(uncom, "tn:") {
				repl = strings.TrimSpace(strings.TrimPrefix(uncom, "tn:"))
			}
			continue
		}
		tn := fmt.Sprintf(TN_BASE, no)
		if repl != "" {
			tn = repl
			repl = ""
		}
		t := NewTask(getTnPrefix() + tn)
		no++
		if err := t.SetCron(ts); err != nil {
			fmt.Fprintf(os.Stderr, "Line %d: %s\n", i, err)
			continue
		}
		if err := t.Register(); err != nil {
			return err
		}
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
	out, err := exec.Command(SCHTASKS_EXE, FLG_QUERY, FLG_FORMAT, FO_CSV, FLG_NOHEADER).Output()
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
		if strings.HasPrefix(r[QUERY_TN_IDX], p) {
			t = append(t, *NewTask(r[QUERY_TN_IDX]))
		}
	}
	return t, nil
}

// Register registers this task at Task Scheduler
func (t *Task) Register() error {
	if t.tn == "" || t.tr == "" || t.sc == "" {
		return errors.New("schedule is unset yet")
	}
	args := []string{FLG_CREATE, FLG_FORCE, FLG_TASKNAME, t.tn, FLG_TASKRUN, t.tr, FLG_SCHEDULE, t.sc}
	if t.st != "" {
		args = append(args, FLG_STARTTIME, t.st)
	}
	if t.mo != "" {
		args = append(args, FLG_MODIFIERS, t.mo)
	}
	if t.d != "" {
		args = append(args, FLG_DAY, t.d)
	}
	if t.m != "" {
		args = append(args, FLG_MONTH, t.m)
	}
	if out, err := exec.Command(SCHTASKS_EXE, args...).CombinedOutput(); err != nil {
		return fmt.Errorf("schtasks runtime error\n  args: %s\n  output: %s", args, string(out))
	}
	return nil
}

// Unregister unregisters this task from Task Scheduler
func (t *Task) Unregister() error {
	if t.tn == "" {
		return errors.New("task name is unset yet")
	}
	args := []string{FLG_DELETE, FLG_FORCE, FLG_TASKNAME, t.tn}
	if out, err := exec.Command(SCHTASKS_EXE, args...).CombinedOutput(); err != nil {
		return fmt.Errorf("schtasks runtime error\n  args: %s\n  output: %s", args, string(out))
	}
	return nil
}

// SetCron sets schedule setting from a row of crontab format text
func (t *Task) SetCron(s string) error {
	f := strings.Fields(s)
	if len(f)-1 < FLD_COMMAND {
		return errors.New("too few fields")
	}
	// Set command
	t.tr = strings.Join(f[FLD_COMMAND:], " ")

	// Parse minute
	m, ms, err := parseField(f[FLD_MINUTE])
	if err != nil {
		return fmt.Errorf("minute has %s", err)
	}
	if len(m) > 1 {
		return errors.New("minute must be single")
	}
	if err := checkRange(m, 0, 59); err != nil {
		return errors.New("minute is out of range")
	}
	if ms < 0 || ms > 1439 {
		return errors.New("minute interval is out of range")
	}

	// Parse hour
	h, hs, err := parseField(f[FLD_HOUR])
	if err != nil {
		return fmt.Errorf("hour has %s", err)
	}
	if len(h) > 1 {
		return errors.New("hour must be single")
	}
	if err := checkRange(h, 0, 23); err != nil {
		return errors.New("hour is out of range")
	}
	if hs < 0 || hs > 23 {
		return errors.New("hour interval is out of range")
	}

	// Parse day of month
	dom, doms, err := parseField(f[FLD_DAY])
	if err != nil {
		return fmt.Errorf("day of month has %s", err)
	}
	if len(dom) > 1 {
		return errors.New("day of month must be single")
	}
	if err := checkRange(dom, 1, 31); err != nil {
		return errors.New("day of month is out of range")
	}
	if doms < 0 || doms > 365 {
		return errors.New("day of month interval is out of range")
	}

	// Parse month
	mon, mons, err := parseField(f[FLD_MONTH])
	if err != nil {
		return fmt.Errorf("month has %s", err)
	}
	if err := checkRange(mon, 1, 12); err != nil {
		return errors.New("month is out of range")
	}
	if mons < 0 || mons > 12 {
		return errors.New("month interval is out of range")
	}

	// Parse day of week
	dow, dows, err := parseField(f[FLD_WEEKDAY])
	if err != nil {
		return fmt.Errorf("day of week has %s", err)
	}
	if err := checkRange(dow, 0, 7); err != nil {
		return errors.New("day of week is out of range")
	}
	if dows < 0 || dows > 52 {
		return errors.New("day of week interval is out of range")
	}

	// Set schedule type
	if m[0] == -1 {
		t.sc = SC_MINUTE
		if ms > 0 {
			t.mo = fmt.Sprintf("%d", ms)
		}
	} else if h[0] == -1 {
		t.sc = SC_HOURLY
		if hs > 0 {
			t.mo = fmt.Sprintf("%d", hs)
		}
	} else if dow[0] != -1 {
		t.sc = SC_WEEKLY
		if dows > 0 {
			t.mo = fmt.Sprintf("%d", dows)
		}
	} else if dom[0] == -1 {
		t.sc = SC_DAILY
		if doms > 0 {
			t.mo = fmt.Sprintf("%d", doms)
		}
	} else {
		t.sc = SC_MONTHLY
		if mons > 0 {
			t.mo = fmt.Sprintf("%d", mons)
		}
		// Set month
		if mon[0] != -1 {
			m := []string{}
			for _, v := range mon {
				m = append(m, itomon(v))
			}
			t.m = strings.Join(m, ",")
		}
	}

	// Set day
	if t.sc == SC_WEEKLY {
		d := []string{}
		for _, v := range dow {
			d = append(d, itodow(v))
		}
		t.d = strings.Join(d, ",")
	} else if t.sc == SC_MONTHLY {
		d := []string{}
		for _, v := range dom {
			d = append(d, fmt.Sprintf("%d", v))
		}
		t.d = strings.Join(d, ",")
	}

	// Set start time
	{
		mm := m[0]
		if mm == -1 {
			mm = 0
		}
		hh := h[0]
		if hh == -1 {
			hh = 0
		}
		t.st = fmt.Sprintf("%02d:%02d", hh, mm)
	}

	return nil
}

func parseField(s string) ([]int, int, error) {
	tmst := strings.Split(s, "/")
	if len(tmst) > 2 {
		return nil, 0, errors.New("too many slashes")
	}
	times := []int{}
	uniq := map[int]struct{}{}
	for _, v := range strings.Split(tmst[0], ",") {
		minmax := strings.Split(v, "-")
		switch len(minmax) {
		case 1:
			i, err := dtstoi(minmax[0])
			if err != nil {
				if len(times) == 0 && minmax[0] == "*" {
					// -1 is every
					times = append(times, -1)
					break
				}
				return nil, 0, err
			}
			if _, ok := uniq[i]; !ok {
				uniq[i] = struct{}{}
				times = append(times, i)
			}
		case 2:
			min, err := dtstoi(minmax[0])
			if err != nil {
				return nil, 0, err
			}
			max, err := dtstoi(minmax[1])
			if err != nil {
				return nil, 0, err
			}
			if min > max {
				return nil, 0, errors.New("min value greater than max")
			}
			for i := min; i <= max; i++ {
				if _, ok := uniq[i]; !ok {
					uniq[i] = struct{}{}
					times = append(times, i)
				}
			}
		default:
			return nil, 0, errors.New("too many hyphens")
		}
	}
	if len(times) == 0 {
		return nil, 0, errors.New("no value")
	}
	step := 0
	if len(tmst) == 2 {
		var err error
		step, err = strconv.Atoi(tmst[1])
		if err != nil {
			return nil, 0, errors.New("invalid step number")
		}
	}
	return times, step, nil
}

func checkRange(nums []int, min int, max int) error {
	for i, v := range nums {
		if i == 0 && v == -1 {
			return nil
		}
		if v < min || v > max {
			return errors.New("out of range")
		}
	}
	return nil
}
