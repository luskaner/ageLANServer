# Launcher Config Admin Agent

The launcher config admin agent is a service designed to avoid repeated admin elevation dialogs executing `config-admin`
while in the background. This agent is started/stopped by `config` as needed.