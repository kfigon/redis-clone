usage:
* open telnet, connect to 6380
    * putty on windows will do
    * `netcat localhost 6380` on linux
* send data

optionally u can use client and send data there

# Protocol
Project uses RESP protocol on top of TCP. Supported commands:
* PING
* ECHO
* GET
* SET
* DELETE