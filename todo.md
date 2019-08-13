# Todo

## Requirements

-   [x] Sniff/Capture Traffic

  -   [x] On all network interfaces
  -   [x] Add filters :
  
    -   [x] TCP and port 80
    
-   [x] Time based console display with traffic information

  -   [x] Most hits on website sections
  
-   [x] Alert when traffic over past n minutes hits a threshold

  -   [x] Inform when recovered

-   [x] Summary statistics

  -   [x] Total number of hits per host
  -   [x] Response statistics per host
  -   [x] Request methods per section
  -   [x] Traffic speed

-   [x] Tests


### Other Todo

-   [x] Configure with command line arguments

-   [ ] When shutting down, if traffic is still incoming, ip adresses are printed to console

-   [x] Tests :

  -   Platforms
  
    -   [x] Linux
    -   [x] Macos
    
  -   With real world traffic
  
    -   [x] Automated crawler
    
  -   Ad-hoc tests

    -   [x] Tailor specific requests to trigger hit detection and alerts

-   [x] Code quality

  -   [x] goreportcard
  -   [x] Codacy
  -   [x] Sonar
  
-   [x] Vulnerability checks

  -   [x] Snyk
