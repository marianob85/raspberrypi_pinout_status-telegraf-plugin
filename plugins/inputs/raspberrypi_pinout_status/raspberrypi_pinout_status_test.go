//go:build linux
// +build linux

package raspberrypi_pinout_status

import (
	"strconv"
	"testing"

	"github.com/influxdata/telegraf/testutil"
	"gotest.tools/assert"
)

var mockSampleReplay = `BANK0 (GPIO 0 to 27):
GPIO 0: level=1 fsel=0 func=INPUT pull=UP
GPIO 1: level=1 fsel=0 func=INPUT pull=UP
GPIO 2: level=1 fsel=0 func=INPUT pull=UP`

var mockFullReplay = `BANK0 (GPIO 0 to 27):
GPIO 0: level=1 fsel=0 func=INPUT pull=UP
GPIO 1: level=1 fsel=0 func=INPUT pull=UP
GPIO 2: level=1 fsel=0 func=INPUT pull=UP
GPIO 3: level=1 fsel=0 func=INPUT pull=UP
GPIO 4: level=1 fsel=0 func=INPUT pull=NONE
GPIO 5: level=1 fsel=0 func=INPUT pull=UP
GPIO 6: level=1 fsel=0 func=INPUT pull=UP
GPIO 7: level=1 fsel=0 func=INPUT pull=UP
GPIO 8: level=1 fsel=0 func=INPUT pull=UP
GPIO 9: level=0 fsel=0 func=INPUT pull=DOWN
GPIO 10: level=0 fsel=0 func=INPUT pull=DOWN
GPIO 11: level=0 fsel=0 func=INPUT pull=DOWN
GPIO 12: level=0 fsel=0 func=INPUT pull=DOWN
GPIO 13: level=0 fsel=0 func=INPUT pull=DOWN
GPIO 14: level=1 fsel=0 func=INPUT pull=NONE
GPIO 15: level=1 fsel=0 func=INPUT pull=UP
GPIO 16: level=0 fsel=0 func=INPUT pull=DOWN
GPIO 17: level=0 fsel=1 func=OUTPUT pull=DOWN
GPIO 18: level=0 fsel=0 func=INPUT pull=DOWN
GPIO 19: level=0 fsel=0 func=INPUT pull=DOWN
GPIO 20: level=0 fsel=0 func=INPUT pull=DOWN
GPIO 21: level=0 fsel=0 func=INPUT pull=DOWN
GPIO 22: level=0 fsel=0 func=INPUT pull=DOWN
GPIO 23: level=0 fsel=0 func=INPUT pull=DOWN
GPIO 24: level=0 fsel=0 func=INPUT pull=DOWN
GPIO 25: level=0 fsel=0 func=INPUT pull=DOWN
GPIO 26: level=0 fsel=0 func=INPUT pull=DOWN
GPIO 27: level=0 fsel=0 func=INPUT pull=DOWN
BANK1 (GPIO 28 to 45):
GPIO 28: level=1 fsel=2 alt=5 func=RGMII_MDIO pull=UP
GPIO 29: level=0 fsel=2 alt=5 func=RGMII_MDC pull=DOWN
GPIO 30: level=0 fsel=7 alt=3 func=CTS0 pull=UP
GPIO 31: level=0 fsel=7 alt=3 func=RTS0 pull=NONE
GPIO 32: level=1 fsel=7 alt=3 func=TXD0 pull=NONE
GPIO 33: level=1 fsel=7 alt=3 func=RXD0 pull=UP
GPIO 34: level=1 fsel=7 alt=3 func=SD1_CLK pull=NONE
GPIO 35: level=1 fsel=7 alt=3 func=SD1_CMD pull=UP
GPIO 36: level=1 fsel=7 alt=3 func=SD1_DAT0 pull=UP
GPIO 37: level=1 fsel=7 alt=3 func=SD1_DAT1 pull=UP
GPIO 38: level=1 fsel=7 alt=3 func=SD1_DAT2 pull=UP
GPIO 39: level=1 fsel=7 alt=3 func=SD1_DAT3 pull=UP
GPIO 40: level=0 fsel=4 alt=0 func=PWM1_0 pull=NONE
GPIO 41: level=0 fsel=4 alt=0 func=PWM1_1 pull=NONE
GPIO 42: level=0 fsel=1 func=OUTPUT pull=UP
GPIO 43: level=1 fsel=0 func=INPUT pull=UP
GPIO 44: level=1 fsel=0 func=INPUT pull=UP
GPIO 45: level=1 fsel=0 func=INPUT pull=UP
BANK2 (GPIO 46 to 53):
GPIO 46: level=0 fsel=0 func=INPUT pull=UP
GPIO 47: level=0 fsel=0 func=INPUT pull=UP
GPIO 48: level=0 fsel=0 func=INPUT pull=DOWN
GPIO 49: level=0 fsel=0 func=INPUT pull=DOWN
GPIO 50: level=0 fsel=0 func=INPUT pull=DOWN
GPIO 51: level=0 fsel=0 func=INPUT pull=DOWN
GPIO 52: level=0 fsel=0 func=INPUT pull=DOWN
GPIO 53: level=0 fsel=0 func=INPUT pull=DOWN`

var mockFullReplayWithParam = `GPIO 0: level=1 fsel=0 func=INPUT pull=UP
GPIO 29: level=0 fsel=2 alt=5 func=RGMII_MDC pull=DOWN
GPIO 38: level=1 fsel=7 alt=3 func=SD1_DAT2 pull=UP`

func funcMockSampleReplay(binary string, args ...string) (string, error) {
	return mockSampleReplay, nil
}

func funcMockFullReplay(binary string, args ...string) (string, error) {
	return mockFullReplay, nil
}

func funcMockFullReplayWithParam(binary string, args ...string) (string, error) {
	return mockFullReplayWithParam, nil
}

var sampleReplayValidate = map[int][]PinOutStatus{
	0: {{0, 1, 0, "INPUT", "UP"},
		{1, 1, 0, "INPUT", "UP"},
		{2, 1, 0, "INPUT", "UP"}},
}

var fullReplayValidate = map[int][]PinOutStatus{
	0: {{0, 1, 0, "INPUT", "UP"},
		{1, 1, 0, "INPUT", "UP"},
		{2, 1, 0, "INPUT", "UP"},
		{3, 1, 0, "INPUT", "UP"},
		{4, 1, 0, "INPUT", "NONE"},
		{5, 1, 0, "INPUT", "UP"},
		{6, 1, 0, "INPUT", "UP"},
		{7, 1, 0, "INPUT", "UP"},
		{8, 1, 0, "INPUT", "UP"},
		{9, 0, 0, "INPUT", "DOWN"},
		{10, 0, 0, "INPUT", "DOWN"},
		{11, 0, 0, "INPUT", "DOWN"},
		{12, 0, 0, "INPUT", "DOWN"},
		{13, 0, 0, "INPUT", "DOWN"},
		{14, 1, 0, "INPUT", "NONE"},
		{15, 1, 0, "INPUT", "UP"},
		{16, 0, 0, "INPUT", "DOWN"},
		{17, 0, 1, "OUTPUT", "DOWN"},
		{18, 0, 0, "INPUT", "DOWN"},
		{19, 0, 0, "INPUT", "DOWN"},
		{20, 0, 0, "INPUT", "DOWN"},
		{21, 0, 0, "INPUT", "DOWN"},
		{22, 0, 0, "INPUT", "DOWN"},
		{23, 0, 0, "INPUT", "DOWN"},
		{24, 0, 0, "INPUT", "DOWN"},
		{25, 0, 0, "INPUT", "DOWN"},
		{26, 0, 0, "INPUT", "DOWN"},
		{27, 0, 0, "INPUT", "DOWN"}},
	1: {{28, 1, 2, "RGMII_MDIO", "UP"},
		{29, 0, 2, "RGMII_MDC", "DOWN"},
		{30, 0, 7, "CTS0", "UP"},
		{31, 0, 7, "RTS0", "NONE"},
		{32, 1, 7, "TXD0", "NONE"},
		{33, 1, 7, "RXD0", "UP"},
		{34, 1, 7, "SD1_CLK", "NONE"},
		{35, 1, 7, "SD1_CMD", "UP"},
		{36, 1, 7, "SD1_DAT0", "UP"},
		{37, 1, 7, "SD1_DAT1", "UP"},
		{38, 1, 7, "SD1_DAT2", "UP"},
		{39, 1, 7, "SD1_DAT3", "UP"},
		{40, 0, 4, "PWM1_0", "NONE"},
		{41, 0, 4, "PWM1_1", "NONE"},
		{42, 0, 1, "OUTPUT", "UP"},
		{43, 1, 0, "INPUT", "UP"},
		{44, 1, 0, "INPUT", "UP"},
		{45, 1, 0, "INPUT", "UP"}},
	2: {{46, 0, 0, "INPUT", "UP"},
		{47, 0, 0, "INPUT", "UP"},
		{48, 0, 0, "INPUT", "DOWN"},
		{49, 0, 0, "INPUT", "DOWN"},
		{50, 0, 0, "INPUT", "DOWN"},
		{51, 0, 0, "INPUT", "DOWN"},
		{52, 0, 0, "INPUT", "DOWN"},
		{53, 0, 0, "INPUT", "DOWN"}},
}

var fullReplayValidateWithParam = map[int][]PinOutStatus{
	-1: {{0, 1, 0, "INPUT", "UP"},
		{29, 0, 2, "RGMII_MDC", "DOWN"},
		{38, 1, 7, "SD1_DAT2", "UP"}},
}

func PerformTest(input map[int][]PinOutStatus, t *testing.T, host HostCommand) {
	var acc testutil.Accumulator
	rpi := RaspberrypiPinoutStatus{
		hostCommand: host,
	}

	acc.GatherError(rpi.Gather)

	for bank, stats := range input {
		for _, gpio := range stats {
			tags := map[string]string{"gpio": strconv.Itoa(gpio.GPIO)}
			if bank >= 0 {
				tags["bank"] = strconv.Itoa(bank)
			}
			fields := map[string]interface{}{
				"level": gpio.Level,
				"fsel":  gpio.FSelect,
				"func":  gpio.Function,
				"pull":  gpio.Pull,
			}
			acc.AssertContainsTaggedFields(t, "rpi_pinout", fields, tags)
		}
	}
}

func TestSampleInput(t *testing.T) {
	PerformTest(sampleReplayValidate, t, funcMockSampleReplay)
}

func TestFullInput(t *testing.T) {
	PerformTest(fullReplayValidate, t, funcMockFullReplay)
}

func TestFullInputWithParameters(t *testing.T) {
	PerformTest(fullReplayValidateWithParam, t, funcMockFullReplayWithParam)
}

func TestDefaultArgs(t *testing.T) {
	rpi := RaspberrypiPinoutStatus{}

	assert.Equal(t, rpi.GetArgs()[0], "get")
}
func TestCustomArgs(t *testing.T) {
	rpi := RaspberrypiPinoutStatus{}

	rpi.GPins = []int{0, 2, 5, 30}

	assert.Equal(t, rpi.GetArgs()[0], "get")
	assert.Equal(t, rpi.GetArgs()[1], "0,2,5,30")
}
