//go:build linux
// +build linux

package raspberrypi_pinout_status

import (
	"bytes"
	"errors"
	"fmt"
	"log"
	"os/exec"
	"regexp"
	"strconv"
	"strings"
	"syscall"
	"time"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/plugins/inputs"
)

type PinOutStatus struct {
	GPIO     int
	Level    int
	FSelect  int
	Function string
	Pull     string
}

type HostCommand func(binary string, args ...string) (string, error)

type RaspberrypiPinoutStatus struct {
	hostCommand HostCommand
	Log         telegraf.Logger `toml:"-"`
	GPins       []int
}

func (rpi *RaspberrypiPinoutStatus) Description() string {
	return "Get pinout status on raspberry pi"
}

const sampleConfig = `
	gpins = [0,1,2,3]
`

func (rpi *RaspberrypiPinoutStatus) parseOutput(input string) (map[int][]PinOutStatus, error) {
	lines := strings.Split(input, "\n")
	reBank := regexp.MustCompile(`^BANK(\d+)`)
	reGpio := regexp.MustCompile(`^GPIO (\d+): level=(\d+) fsel=(\d+).*func=([0-9A-Za-z_]+) pull=([0-9A-Za-z]+)`)
	bank := -1
	ret := make(map[int][]PinOutStatus)
	for _, line := range lines {
		if digit := reBank.FindStringSubmatch(line); digit != nil {
			val, err := strconv.Atoi(digit[1])
			if err != nil {
				return nil, err
			}
			bank = val
		} else if stat := reGpio.FindStringSubmatch(line); stat != nil {
			var pinout PinOutStatus
			var err error
			pinout.GPIO, err = strconv.Atoi(stat[1])
			if err != nil {
				return nil, err
			}
			pinout.Level, err = strconv.Atoi(stat[2])
			if err != nil {
				return nil, err
			}
			pinout.FSelect, err = strconv.Atoi(stat[3])
			if err != nil {
				return nil, err
			}
			pinout.Function = stat[4]
			pinout.Pull = stat[5]

			ret[bank] = append(ret[bank], pinout)
		}
	}
	return ret, nil
}

func (rpi *RaspberrypiPinoutStatus) SampleConfig() string {
	return sampleConfig
}

func (rpi *RaspberrypiPinoutStatus) GetArgs() []string {
	var args = []string{"get"}
	if rpi.GPins != nil {
		args = append(args, strings.Trim(strings.Join(strings.Fields(fmt.Sprint(rpi.GPins)), ","), "[]"))
	}
	return args
}

func (rpi *RaspberrypiPinoutStatus) Gather(acc telegraf.Accumulator) error {
	data, err := rpi.hostCommand("raspi-gpio", rpi.GetArgs()...)

	if err != nil {
		rpi.Log.Errorf("raspi-gpio failed: %s", err.Error())
		return err
	}

	pinout, err := rpi.parseOutput(data)

	if err != nil {
		return err
	}
	for bank, stats := range pinout {
		for _, gpio := range stats {
			tags := map[string]string{"gpio": strconv.Itoa(int(gpio.GPIO))}
			if bank >= 0 {
				tags["bank"] = strconv.Itoa(bank)
			}

			fields := map[string]interface{}{
				"level": gpio.Level,
				"fsel":  gpio.FSelect,
				"func":  gpio.Function,
				"pull":  gpio.Pull,
			}
			acc.AddFields("rpi_pinout", fields, tags)
		}
	}

	return nil
}

func CombinedOutputTimeout(c *exec.Cmd, timeout time.Duration) ([]byte, error) {
	var b bytes.Buffer
	c.Stdout = &b
	c.Stderr = &b
	if err := c.Start(); err != nil {
		return nil, err
	}
	err := WaitTimeout(c, timeout)
	return b.Bytes(), err
}

// KillGrace is the amount of time we allow a process to shutdown before
// sending a SIGKILL.
const KillGrace = 5 * time.Second

// WaitTimeout waits for the given command to finish with a timeout.
// It assumes the command has already been started.
// If the command times out, it attempts to kill the process.
func WaitTimeout(c *exec.Cmd, timeout time.Duration) error {
	var kill *time.Timer
	term := time.AfterFunc(timeout, func() {
		err := c.Process.Signal(syscall.SIGTERM)
		if err != nil {
			log.Printf("E! [agent] Error terminating process: %s", err)
			return
		}

		kill = time.AfterFunc(KillGrace, func() {
			err := c.Process.Kill()
			if err != nil {
				log.Printf("E! [agent] Error killing process: %s", err)
				return
			}
		})
	})

	err := c.Wait()

	// Shutdown all timers
	if kill != nil {
		kill.Stop()
	}
	termSent := !term.Stop()

	// If the process exited without error treat it as success.  This allows a
	// process to do a clean shutdown on signal.
	if err == nil {
		return nil
	}

	// If SIGTERM was sent then treat any process error as a timeout.
	if termSent {
		return errors.New("command timed out")
	}

	// Otherwise there was an error unrelated to termination.
	return err
}

func hostCommandExecute(binary string, args ...string) (string, error) {
	bin, err := exec.LookPath(binary)
	if err != nil {
		return "", err
	}
	c := exec.Command(bin, args...)

	out, err := CombinedOutputTimeout(c, time.Second*time.Duration(5))
	return string(out), err
}

func init() {
	inputs.Add("raspberrypi_pinout_status", func() telegraf.Input {
		return &RaspberrypiPinoutStatus{
			hostCommand: hostCommandExecute,
		}
	})
}
