# GoHole

A fork of [GoHole](https://github.com/segura2010/GoHole) for the raspberypi 
GoHole is a DNS server written in Golang with the same idea than the [PiHole](https://pi-hole.net), blocking advertisements's and tracking's domains.

The use of sql-lite as the query DB has been replaced with bolt DB and the use of Redis DB as a cache has been replaced with go-cache
This allows a statically linked binary to be produced that builds into a small 6M docker container built using [resinio](https://resin.io/) 


### Usage

To start the DNS server you have to run the following command:

`gohole -s`

You can specify a config file with the command line argument `-c`. See the `config_example.json` file to see the structure. 

You can also provide the `-p` argument to specify the port in which the DNS server will listen.

You can use the secure DNS server generating an AES encryption key using the command "gohole -gkey". Then, download it in your device and configure the [GoHole CryptClient](https://github.com/segura2010/GoHole-CryptClient).

To block ads domains, you must add them to the cache DB. In order to do that, you must pass a blocklist file using the following command:

`gohole -ab path/to/blacklist_file`

If the list is published in a web server, you can provide the URL: 

`gohole -ab http://domain/path/to/blacklist_file`

You can follow this link to get an updated list of available block content:
https://github.com/StevenBlack/hosts

If you does not know any blacklist, you can see the file `blacklists/list.txt`. It contains the blacklists used by the PiHole. You can use a file with a list of blacklist like the `blacklists/list.txt` file to automatically add all the lists:

`gohole -abl blacklists/list.txt`

You can also block domains by using the following command:

`gohole -ad google.com -ip4 0.0.0.0 -ip6 "::1"`

and unblock domains by using the following command:

`gohole -dd google.com`

#### Flush cache and logs

You can flush cache and logs DBs.

**Flush domains cache**

`gohole -fcache`

**Flush logs**

`gohole -flog`


#### Statistics and Logs

You can see the stats and logs by using the following command line arguments:

**See all the clients that have made a request**

`gohole -lc`

**See all request made by a client**

`gohole -lip <clientip>`

**See all clients that queried a domain**

`gohole -ld <clientip>`

**See top domains and number of queries for them**

`gohole -ldt -limit 10`

### Docker

You can use GoHole in a Docker container. 
A DockerFile is availble in the src code, this builds a docker image for raspberypi 
I'm using [resinio](https://resin.io/) to build and manage the containers.

### Metrics & Statistics on Graphite

You can send the statistics to your Graphite server. Configure it on your config file (host and port of the server). Then, you will be able to see your graphs in the Graphite web panel or in Grafana.

![Grafana Dashboard](http://i.imgur.com/6eK98At.png)

You can export the Grafana dashboard I used in the image using the [grafana/GoHole.json](https://github.com/segura2010/GoHole/tree/master/grafana/GoHole.json) file.

You can use the Graphite+Grafana Docker image from: https://github.com/kamon-io/docker-grafana-graphite


**Tested on Go 1.8.3 and 1.9.2**


git push resin rpi-docker:master
resin build --logs --nocache --deviceType raspberrypi3 --arch armhf
sudo ./gohole -s -c ../config_example.json -abl ../blacklists/list.txt -gkey
docker build -t rpassmore/gohole ./
docker run -ti rpassmore/gohole:latest
