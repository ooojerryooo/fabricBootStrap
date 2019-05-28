#!/bin/bash

export FABRIC_CFG_PATH=${PWD}
peer node start >> log_peer0-org1.log 2>&1 &