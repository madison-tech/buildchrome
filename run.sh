#!/bin/zsh
ssh -o StrictHostKeychecking=no $1 screen -dm 'bash -c "cd;
git clone https://github.com/madison-tech/docker-headless-shell.git;
cd docker-headless-shell;
git checkout arm64; 
./build-headless-shell.sh;"' # Docker needs to be installed to run ./build-docker.sh