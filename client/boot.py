import network
import machine


wlan = network.WLAN(network.STA_IF)
while not wlan.isconnected():
    status = wlan.status()
    if status == network.STAT_CONNECTING:
        machine.idle()
    else:
        print("Network status:", status)
        break

