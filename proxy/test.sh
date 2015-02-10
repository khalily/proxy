#!/bin/bash

port=$1

(echo -en "GET / HTTP/1.1\r\nHost: 127.0.0.1\r\n\r\nGET / HTTP/1.1\r\nHost: 127.0.0.1\r\n\r\nGET / HTTP/1.1\r\nHost: 127.0.0.1\r\n\r\n"; sleep 3) | nc 127.0.0.1 $port
