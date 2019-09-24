import livereload

LIVERELOAD_SERVER = "10.0.0.20"
WIFI_SSID = ""
WIFI_PASSWORD = ""

if WIFI_SSID:
    import network

    wlan = network.WLAN(network.STA_IF)
    wlan.connect(WIFI_SSID, WIFI_PASSWORD)

livereload.wait_for_network()
livereload.connect(LIVERELOAD_SERVER, 1808)

# import a script from the livereload server
import main
