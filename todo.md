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
  
-   [ ] Summary statistics

  -   [x] Total number of hits per host
  -   [x] Response statistics per host
  -   [x] Request methods per section
  -   [ ] Traffic speed

-   [x] Tests

### Remain


### Error handling

-  Thoroughly check if all errors are handled

### Other Todo

-   [ ] Tests :

  -   Platforms
  
    -   [x] Linux
    -   [x] Macos
    -   [ ] maybe Windows ?
    
  -   With real world traffic
  
    -   [x] Automated crawler
    
  -   Ad-hoc tests
  
    -   [ ] Tailor specific requests to trigger hit detection and alerts

-   [x] Code quality

  -   [x] goreportcard
  -   [x] Codacy
  -   [x] Sonar
  
-   [x] Vulnerability checks

  -   [x] Snyk
