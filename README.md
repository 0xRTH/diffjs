# diffjs

This tools fetch a list of urls and compare them to a folder of older files.\
It returns new files and files that changed. \
It can be used to monitor any kind of urls but most likely js files to view codes changes. 

## Install

`go install github.com/0xRTH/diffjs@latest`

## Usage : 

### Basic:

`cat urls.txt | diffjs`

### With Notificaton

`cat urls.txt | diffjs -notify`

## Help : 

- -dir Directory containing js files created by this tools previously to compare (default: "./Downloads") 
- -log Directory to output the logs to (default: "./logs") 
- -archive Directory where to store old files (default: "./Archives") 
- -notify activate notification (notify is needed)
- -s Silent
- -v Get all infos in the output
- -h Show help

## Todos

- Output the diff if it's short enough
- Integrate Notify directly in go so it's not up to the user to install it separately

## Requirements

- Golang v1.18
- [Notify](https://github.com/projectdiscovery/notify) by projectdiscovery
