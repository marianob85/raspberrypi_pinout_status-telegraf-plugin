# RaspberrypiPinoutStatus Input Plugin

This input plugin will gather rapsberry pi gpio state

### Install Instructions 

To integrate with telegraf, extend the telegraf.conf using the following example
```
[[inputs.execd]]
   command = ["raspberrypi_pinout_status-telegraf-plugin", "-config", "/etc/telegraf/raspberrypi_pinout_status.conf"]
   signal = "STDIN"
```

### Configuration:

```
[[inputs.raspberrypi_pinout_status]]
	gpins = [17] # optional
```