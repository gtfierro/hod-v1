#!/bin/bash

/bin/hod load -c /etc/hod/hodconfig.yaml /etc/hod/building.ttl
/bin/hod server -c /etc/hod/hodconfig.yaml
