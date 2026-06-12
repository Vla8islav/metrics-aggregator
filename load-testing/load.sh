#!/bin/sh

vegeta attack -targets=attack.txt -duration=90s -rate=150/s | vegeta report