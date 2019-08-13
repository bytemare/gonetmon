# Tracker

> A non-exhaustive list of things done, former to-do list.

## Requirements

- [x] Sniff/Capture Traffic
  - [x] On all network interfaces
  - [x] Add filters :
    - [x] TCP and port 80

- [x] Time based console display with traffic information
  - [x] Display sections of website with the most hits
  
- [x] Alert when traffic over past n minutes hits a threshold
  - [x] Inform when recovered

- [x] Summary statistics
  - [x] Total number of hits per host
  - [x] Response statistics per host
  - [x] Request methods per section
  - [x] Traffic speed

- [x] Tests

## Other

- [x] Configure timeout with command line argument
- [x] Tests :
  - Platforms
    - [x] Linux
    - [ ] Macos
  - With real world traffic
    - [x] Automated crawler
- [x] Code quality
  - [x] goreportcard
  - [x] Codacy
  - [x] Sonar
- [x] Vulnerability checks
  - [x] Snyk
