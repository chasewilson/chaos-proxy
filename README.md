# chaos-proxy
Build a simple **TCP proxy** in Go that forwards connections based on the **local port** they arrive on.   Each port maps to a different upstream target and can optionally inject network faults such as latency, drops, or timeouts.
