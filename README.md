This README is work in progress.
# pingd
Daemon written in Golang to provide a microservice that pings a given address (ipv4 only in this version) and returns the response time in milliseconds as output.
Second argument is the timeout on the ping request in seconds.
Third argument is the number of pings.

Use case for the daemon is to allow for Zabbix to have an easy "ping" service that does not require any additional rights or tools to work and can be deployed on any Linux platform.
No Windows version available, only Linux.
