# Todo

## Requirements

- [x] Sniff/Capture Traffic
    - [x] On all network interfaces
    - [x] Add filters :
        - [x] TCP and port 80
- [x] Time based console display with traffic information
    - [ ] Most hits on website sections
- [ ] Alert when traffic over past n minutes hits a threshold
    - [ ] Inform when recovered
- [ ] Tests

### Remain

- Alert logic
    - Time based LRU cache of hits
    - dedicated go routine with ticker to evict elements : but what about concurrency accessing/modifying the elements ?
        - go through first elements and discard every element that has expired
    - hits on timespan = number of element in the list
    - at each tick, verify if alert is triggered : if there's a change of state, alert, if recovering, inform.
        - if we were already in alert, don't alert again or it will flood the display
- Proper synchronisation for graceful shutdown
- Clean display and colors


## Error handling

- Thoroughly check if all errors are handled 

## Other Todo

- [ ] Tests :
    - Platforms
        - [ ] Different Linux platforms
        - [ ] Macos
        - [ ] maybe Windows ?
    - With real world traffic
        - [ ] Simulate with user behaviour ?
        - [ ] Automated crawler ?
    - Ad-hoc tests
        - [ ] Tailor specific requests to trigger hit detection and alerts

- [ ] Code quality
    - 
    - [ ] Codacy
    - [ ] Sonar
- [ ] Vulnerability checks
