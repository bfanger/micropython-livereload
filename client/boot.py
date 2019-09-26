import livereload

# livereload.wifi("YOUR_SSID", "YOUR_PASSWORD")
livereload.wait_for_network()
livereload.connect("YOUR_IP", 1808)
livereload.detect(500)

# uncomment for unix port (boards will `import main` automaticly after boot.py)
# import main
