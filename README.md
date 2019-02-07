# Payroll Web Application

## Pre-req

1. Download and install docker and docker-compose.
2. Download and install Go. For this, you need to set up $GOPATH, $GOROOT and \$GOBIN.
3. Download and install npm.
4. Port 80 should be available for the application to serve.

## Steps to run the application

1. Place the provides tarball under path: go/src/github.com/ravjotsingh9/payroll-application
2. Make sure you have docker, docker-compose installed and have access to internet.
3. CD to payroll-application dir and run `docker-compose up --build`. This should start 4 docker processes.
4. On another terminal, CD payroll-application/frontend and run `npm install`
5. After npm install is done, start the frontend by running `npm start`. This should start frontend on http://localhost:3000
