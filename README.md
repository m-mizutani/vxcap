# vxcap

[![Travis-CI](https://travis-ci.org/m-mizutani/vxcap.svg)](https://travis-ci.org/m-mizutani/vxcap) [![Report card](https://goreportcard.com/badge/github.com/m-mizutani/vxcap)](https://goreportcard.com/report/github.com/m-mizutani/vxcap)

Capture and dump VXLAN encapsulated traffic. Main focus is AWS VPC traffic mirroring.

![arch](https://user-images.githubusercontent.com/605953/64929961-06461c80-d867-11e9-8f83-841c94b84c85.png)

## Setup

### Prerequisite

- Go >= 1.11.1

### Install

```bash
go install github.com/m-mizutani/vxcap
```

## Getting started

### Capture traffic and save packet to file as pcap format

```bash
vxcap -f pcap -e fs --fs-filename your_dump_file.pcap
```

### Capture traffic and save packet to AWS S3 Bucket as json record

```bash
vxcap -f json -e s3 --aws-region ap-northeast-1 --aws-s3-bucket your-bucket-name
```

### Capture traffic and send packet data to AWS Firehose

```bash
vxcap -f json -e firehose --aws-region ap-northeast-1 --aws-firehose-name your-hose-name
```

## Options

- Base options
  - `--emitter <value>, -e <value>`:  Destination to save data [fs,s3,firehose] (default: "fs")
  - `--dumper <value>, -d <value>`:  Write format [pcap,json] (default: "pcap")
  - `--log-level <value>`:  Log level [trace,debug,info,warn,error] (default: "info")
- Options for UDP server to receive VXLAN packet
  - `--port <value>, -p <value>`:  UDP port of VXLAN receiver (default: 4789)
  - `--receiver-queue-size <value>`:  Queue size between UDP server and packet processor (default: 1024)
- Options for file system emitter (`fs`)
  - `--fs-filename <value>`:  Base file name for FS emitter (default: "dump")
  - `--fs-dirpath <value>`:  Output directory for FS emitter (default: ".")
- Options for AWS service emitter (`s3` and `firehose`)
  - `--aws-region <value>`:  AWS region for emitter to AWS
  - `--aws-s3-bucket <value>`:  AWS S3 bucket name for S3 emitter
  - `--aws-s3-prefix <value>`:  Prefix of AWS S3 object key for S3 emitter
  - `--aws-s3-add-time-key`:  Enable to add time key to S3 object key for S3 emitter
  - `--aws-s3-flush-count <value>`:  Threshold of record number to flush object to AWS S3 bucket
  - `--aws-s3-flush-interval <value>`: Flush interval (seconds) to AWS S3 bucket
  - `--aws-firehose-name <value>`:  Name of AWS Firehose for Firehose emitter
  - `--aws-firehose-flush-size <value>`  Threshold of record size to flush object to AWS Firehose
  - `--aws-firehose-flush-interval <value>`: Flush interval (seconds) to AWS Firehose
- Options for JSON format
  - `--enable-json-text`:  Enable human readable application layer payload in json format
  - `--enable-json-raw`:  Enable raw application layer payload (base64 encoded) in json format

## Test

```bash
go test ./...
```

## Author and License

- Author: Masayoshi Mizutani mizutani@sfc.wide.ad.jp / [@m_mizutani](https://twitter.com/m_mizutani)
- [MIT License](./LICENSE)
