# Redis clone

Toy project that imitates Redis. Uses actual RESP and inmemory golang's map as a storage.

## Usage:
* open telnet, connect to 6380
    * putty on windows will do
    * `netcat localhost 6380` on linux
* send data

You can also use built in CLI. It allows you to run in the following modes. Run with `-h` to learn more
* server
* tcp client
* redis client
* stres test

# Protocol
Project uses RESP protocol on top of TCP. Supported commands:
* PING
* ECHO
* GET
* SET
* DELETE