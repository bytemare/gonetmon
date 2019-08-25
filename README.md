# gonetmon

> A network activity monitor in Go.

<p align="center">

[![Build Status](https://travis-ci.com/bytemare/gonetmon.svg?branch=master)](https://travis-ci.com/bytemare/gonetmon)
[![Go Report Card](https://goreportcard.com/badge/github.com/bytemare/gonetmon)](https://goreportcard.com/report/github.com/bytemare/gonetmon)
[![codebeat badge](https://codebeat.co/badges/4b68d6e5-0333-441d-9964-e297530097c2)](https://codebeat.co/projects/github-com-bytemare-gonetmon-master) 
[![Codacy Badge](https://api.codacy.com/project/badge/Grade/5bc1136110874ceab9195a31bb0e3961)](https://www.codacy.com/app/bytemare/gonetmon)
[![GoDoc](https://godoc.org/github.com/bytemare/gonetmon?status.svg)](https://godoc.org/github.com/bytemare/gonetmon)
</p>

The gonetmon package captures HTTP traffic and displays interesting statistics about the traffic.

## Installing / Getting started

> ## Notes
> 
> For now, gonetmon reliably works only on Linux. gopacket panics on MacOS and Windows support has not yet been integrated.

In order to be able to capture packets, you'll need the libpcap library. On your favorite Linux distribution, install it like so :

```shell
sudo apt-get install libpcap-dev
```

This will install libpcap-dev on your machine an allow you to capture packets / sniff traffic.

Let's suppose you have [a working environment for Go](https://golang.org/doc/install). All that's left to do is getting the package :

```shell
go get github.com/bytemare/gonetmon
```

That just downloaded the project into your $GOPATH/src, and you're set.

## 'Alright, get me to it !'

Here you go :

```shell
cd $GOPATH/src/github.com/bytemare/gonetmon/Tests
go build sniffer.go
sudo ./sniffer
```

We need to run with elevated privileges, since the system wouldn't let us capture packets otherwise.
This will clear your terminal and start showing things like the current http traffic, speed, top visited site, and even show some alerts if the traffic is high.

Not seeing anything ? That's maybe because there's no traffic, or because it's encrypted. Reminder : this only shows plaintext HTTP traffic.
But don't worry, I got your back ! On the same machine, open another terminal :

```shell
cd Tests/RealTraffic/
go get
go run RealTraffic.go &> /dev/null
```

This is a webcrawler that will generate a lot of plaintext traffic for a minute or so, exactly what we need !

This is the kind of output that we'll have :

![Image1](/images/img1.png)

Note that the traffic spike triggered an alert. After some moments, when the storm is down, we'll have a message that we recovered from alert.

![Image2](/images/img2.png)

A handy little option for our sniffer here is that it can take a timeout (in seconds) as an argument, and will close itself after that timeout. Interesting if you want to dump your traffic for some time without being there.

```shell
sudo ./sniffer -timeout=200
```

In every case, you can gracefully shut down the monitoring by gently hitting CTRL+C on your keyboard.

## Configuration

For now all configuration parameters have default values in the code. But it is fairly easy to change them in order to change the programs behaviour, just take a look a [params.go](https://github.com/bytemare/gonetmon/blob/master/params.go).

## Documentation

If you want to use specific functions, please read up on them in the [documentation](https://godoc.org/github.com/bytemare/gonetmon).

## Todo

Like all engineering projects, there's always room to do better, and these are some of the next things I want to do :

### Corrections

- Improve documentation and its layout
- When shutting down, the collector continues logging received packets' IP addresses. That must have something to do with messages still in the PacketSource channel. It would be better if this wouldn't happen.
- Proper 'init()' functions that takes profit of go's 'init()' interpretation

### Features

- Ability to fully configure program behaviour with command line arguments and configuration file
- Richer logging
- Add more and better logs
- Make it work on MacOS
- Make it work on Windows
- during runtime, continually watch out for new devices being opened
- export results to different formats : json and/or html to display it in a browser ?
- TCP Stream reassembly : coherently reassemble packets and calculate connection quality based upon round-trips
- Ability to add more filters
