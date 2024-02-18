# diffjs

## Install

`go install github.com/0xRTH/diffjs@latest`

## Usage : 

### Basic:
`cat urls.txt | diffjs`

### With Notificaton

## Help : 

- -dir Directory containing js files created by this tools previously to compare (default: "./Downloads") 
- -log Directory to output the logs to (default: "./logs") 
- -archive Directory where to store old files (default: "./Archives") 
- -notify activate notification (notify is needed)
- -s Silent
- -v Get all infos in the output
- -h Show help

## Requirements

- Golang v1.18
- [Notify](https://github.com/projectdiscovery/notify) by projectdiscovery
