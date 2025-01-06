#!/bin/bash

$(clear)

echo 'Updating Termux packages...'
$(pkg update && pkg upgrade -y)
echo 'Termux packages are updated!'

echo 'Installing Golang package...'
$(pkg install -y golang)
echo 'Golang package installed!'

echo 'Adding environment variables to .bashrc ...'
echo 'export GOPATH=$HOME/go' >> ~/.bashrc
echo 'export PATH=$PATH:$GOPATH/bin' >> ~/.bashrc
echo 'Environment variables added!'

echo 'Reloading bash...'
$(source ~/.bashrc)
echo 'Bash reloaded!'

echo 'Installing mdx application...'
$(go install github.com/arimatakao/mdx@latest)
echo 'mdx installed!'

echo 'Execute the "mdx" command to use the program!'
